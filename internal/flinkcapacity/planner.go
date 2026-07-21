package flinkcapacity

import (
	"fmt"
	"sort"
)

type TopologyError struct {
	message   string
	retryable bool
}

func (e *TopologyError) Error() string   { return e.message }
func (e *TopologyError) Retryable() bool { return e.retryable }

type Action string

const (
	ModifyWorkspaceFixed    Action = "modify_workspace_fixed"
	ModifyWorkspacePostpaid Action = "modify_workspace_postpaid"
	EnableWorkspaceElastic  Action = "enable_workspace_elastic"
	ModifyWorkspaceElastic  Action = "modify_workspace_elastic"
	CreateNamespace         Action = "create_namespace"
	DeleteNamespace         Action = "delete_namespace"
	ModifyNamespace         Action = "modify_namespace"
	ModifyQueue             Action = "modify_queue"
)

const minimumNamespaceCU CU = 2 // one CU, represented in half-CU units.

type Ref struct {
	Namespace string
	Queue     string
}

type Allocation struct {
	FixedCU          CU
	CrossZoneFixedCU CU
	Limit            CU
}

func (a Allocation) TotalFixed() CU {
	return a.FixedCU + a.CrossZoneFixedCU
}

func (a Allocation) AsCapacity() Capacity {
	return Capacity{Fixed: a.TotalFixed(), Limit: a.Limit}
}

type Step struct {
	Action             Action
	Ref                Ref
	From               Allocation
	To                 Allocation
	NamespaceCrossZone bool
	Temporary          bool
}

type candidate struct {
	step      Step
	level     int
	expansion bool
}

// ValidateDesired verifies the authoritative end state. Namespace allocations
// must consume the workspace exactly, and each namespace owns only the
// service-created default queue with the same allocation.
func ValidateDesired(tree Tree) error {
	if err := validateWorkspace(tree.ChargeType, tree.Workspace); err != nil {
		return err
	}
	if err := validateAuthoritativeWorkspaceMode(tree.Workspace); err != nil {
		return err
	}
	if len(tree.Namespaces) == 0 {
		return fmt.Errorf("workspace must declare at least one namespace")
	}

	seen := make(map[string]struct{}, len(tree.Namespaces))
	var fixed, crossZoneFixed, limit CU
	for _, namespace := range tree.Namespaces {
		if namespace.Name == "" {
			return fmt.Errorf("namespace name must not be empty")
		}
		if _, exists := seen[namespace.Name]; exists {
			return fmt.Errorf("workspace has duplicate namespace %q", namespace.Name)
		}
		seen[namespace.Name] = struct{}{}
		if namespace.Capacity == nil {
			return fmt.Errorf("namespace %q capacity is required", namespace.Name)
		}
		if err := namespace.Capacity.Validate(); err != nil {
			return fmt.Errorf("namespace %q: %w", namespace.Name, err)
		}
		if namespace.CrossZone != tree.Workspace.HA {
			return fmt.Errorf("namespace %q cross-zone type does not match workspace fixed CU pool", namespace.Name)
		}
		if namespace.Capacity.Fixed < minimumNamespaceCU {
			return fmt.Errorf("namespace %q fixed CU %v must be at least %v", namespace.Name, namespace.Capacity.Fixed.Float64(), minimumNamespaceCU.Float64())
		}
		if err := validateAuthoritativeNamespace(namespace); err != nil {
			return err
		}
		if err := validateUsedCU(fmt.Sprintf("namespace %q", namespace.Name), namespace.Used, namespace.Capacity.Limit); err != nil {
			return err
		}
		if len(namespace.Queues) != 1 || namespace.Queues[0].Name != "default-queue" {
			return fmt.Errorf("namespace %q must declare only implicit queue \"default-queue\"", namespace.Name)
		}
		queue := namespace.Queues[0]
		if queue.Capacity == nil {
			return fmt.Errorf("queue %q/default-queue capacity is required", namespace.Name)
		}
		if *queue.Capacity != *namespace.Capacity {
			return fmt.Errorf("queue %q/default-queue capacity must equal its namespace capacity", namespace.Name)
		}
		if err := validateUsedCU(fmt.Sprintf("queue %q/default-queue", namespace.Name), queue.Used, queue.Capacity.Limit); err != nil {
			return err
		}
		if namespace.CrossZone {
			crossZoneFixed += namespace.Capacity.Fixed
		} else {
			fixed += namespace.Capacity.Fixed
		}
		limit += namespace.Capacity.Limit
	}
	if err := validateAuthoritativeWorkspaceIntegers(tree.Workspace); err != nil {
		return err
	}
	if fixed != tree.Workspace.FixedCU {
		return fmt.Errorf("namespace fixed CU sum %v must equal workspace fixed CU %v", fixed.Float64(), tree.Workspace.FixedCU.Float64())
	}
	if crossZoneFixed != tree.Workspace.CrossZoneFixedCU {
		return fmt.Errorf("namespace cross-zone fixed CU sum %v must equal workspace cross-zone fixed CU %v", crossZoneFixed.Float64(), tree.Workspace.CrossZoneFixedCU.Float64())
	}
	if limit != tree.Workspace.Limit {
		return fmt.Errorf("namespace max CU sum %v must equal workspace max CU %v", limit.Float64(), tree.Workspace.Limit.Float64())
	}
	return nil
}

// Plan preserves the legacy coordinator contract: topology must already
// exist, arbitrary queues are supported, and one omitted capacity at each
// level receives the parent's remainder through Resolve.
func Plan(actual, desired Tree) ([]Step, error) {
	if actual.ChargeType != desired.ChargeType {
		return nil, fmt.Errorf("workspace charge type cannot change from %q to %q", actual.ChargeType, desired.ChargeType)
	}
	if actual.Workspace.HA != desired.Workspace.HA {
		return nil, fmt.Errorf("workspace high availability mode cannot change in place")
	}
	if err := validateLegacySameTopology(actual, desired); err != nil {
		return nil, err
	}

	state, err := Resolve(actual)
	if err != nil {
		return nil, fmt.Errorf("actual capacity tree: %w", err)
	}
	target, err := Resolve(desired)
	if err != nil {
		return nil, fmt.Errorf("desired capacity tree: %w", err)
	}
	if err := validateLegacyUsedCapacity(state, target); err != nil {
		return nil, err
	}
	actualElastic := state.Workspace.AsCapacity().Elastic()
	desiredElastic := target.Workspace.AsCapacity().Elastic()
	if state.ChargeType == "PRE" && actualElastic > 0 && desiredElastic == 0 {
		return nil, fmt.Errorf("cannot reduce workspace elastic CU to zero through the supported public API; disable it manually before retrying")
	}

	maxEndpointLimit := state.Workspace.Limit
	if target.Workspace.Limit > maxEndpointLimit {
		maxEndpointLimit = target.Workspace.Limit
	}
	var steps []Step
	for attempts := 0; !legacySameCapacityTree(state, target); attempts++ {
		if attempts > capacityNodeCount(target)*4+8 {
			return nil, fmt.Errorf("capacity planner did not converge")
		}
		candidates := buildLegacyCandidates(state, target)
		selected := selectLegacyCandidate(state, candidates, maxEndpointLimit, true)
		if selected == nil {
			selected = selectLegacyCandidate(state, candidates, maxEndpointLimit, false)
		}
		if selected == nil {
			return nil, fmt.Errorf("no safe capacity transition exists without exceeding endpoint capacity or violating child allocations")
		}
		applyCandidate(&state, *selected)
		steps = append(steps, selected.step)
	}
	return steps, nil
}

func validateLegacySameTopology(actual, desired Tree) error {
	desiredNamespaces := make(map[string]Namespace, len(desired.Namespaces))
	for _, namespace := range desired.Namespaces {
		desiredNamespaces[namespace.Name] = namespace
	}
	actualNamespaces := make(map[string]Namespace, len(actual.Namespaces))
	for _, namespace := range actual.Namespaces {
		actualNamespaces[namespace.Name] = namespace
		desiredNamespace, ok := desiredNamespaces[namespace.Name]
		if !ok {
			return &TopologyError{message: fmt.Sprintf("cloud namespace %q is undeclared", namespace.Name)}
		}
		desiredQueues := make(map[string]struct{}, len(desiredNamespace.Queues))
		for _, queue := range desiredNamespace.Queues {
			desiredQueues[queue.Name] = struct{}{}
		}
		for _, queue := range namespace.Queues {
			if _, ok := desiredQueues[queue.Name]; !ok {
				return &TopologyError{message: fmt.Sprintf("cloud queue %q/%q is undeclared", namespace.Name, queue.Name)}
			}
		}
	}
	for _, namespace := range desired.Namespaces {
		actualNamespace, ok := actualNamespaces[namespace.Name]
		if !ok {
			return &TopologyError{message: fmt.Sprintf("declared namespace %q does not exist in the workspace", namespace.Name), retryable: true}
		}
		actualQueues := make(map[string]struct{}, len(actualNamespace.Queues))
		for _, queue := range actualNamespace.Queues {
			actualQueues[queue.Name] = struct{}{}
		}
		for _, queue := range namespace.Queues {
			if _, ok := actualQueues[queue.Name]; !ok {
				return &TopologyError{message: fmt.Sprintf("declared queue %q/%q does not exist in the workspace", namespace.Name, queue.Name), retryable: true}
			}
		}
	}
	return nil
}

func validateLegacyUsedCapacity(actual, desired Tree) error {
	if actual.Workspace.Used > desired.Workspace.Limit.Float64() {
		return fmt.Errorf("workspace desired limit %v is below used CU %v", desired.Workspace.Limit.Float64(), actual.Workspace.Used)
	}
	for _, wanted := range desired.Namespaces {
		observed := namespaceByName(actual, wanted.Name)
		if observed.Used > wanted.Capacity.Limit.Float64() {
			return fmt.Errorf("namespace %q desired limit %v is below used CU %v", wanted.Name, wanted.Capacity.Limit.Float64(), observed.Used)
		}
		for _, wantedQueue := range wanted.Queues {
			observedQueue := queueByName(*observed, wantedQueue.Name)
			if observedQueue.Used > wantedQueue.Capacity.Limit.Float64() {
				return fmt.Errorf("queue %q/%q desired limit %v is below used CU %v", wanted.Name, wantedQueue.Name, wantedQueue.Capacity.Limit.Float64(), observedQueue.Used)
			}
		}
	}
	return nil
}

func buildLegacyCandidates(state, target Tree) []candidate {
	result := make([]candidate, 0, capacityNodeCount(target)+1)
	result = append(result, buildWorkspaceCandidates(state, target)...)
	for _, wanted := range target.Namespaces {
		observed := namespaceByName(state, wanted.Name)
		from := capacityAllocation(*observed.Capacity)
		to := capacityAllocation(*wanted.Capacity)
		if from != to {
			result = append(result, newCandidate(ModifyNamespace, Ref{Namespace: wanted.Name}, from, to, 1))
		}
		for _, wantedQueue := range wanted.Queues {
			observedQueue := queueByName(*observed, wantedQueue.Name)
			from = capacityAllocation(*observedQueue.Capacity)
			to = capacityAllocation(*wantedQueue.Capacity)
			if from != to {
				result = append(result, newCandidate(ModifyQueue, Ref{Namespace: wanted.Name, Queue: wantedQueue.Name}, from, to, 2))
			}
		}
	}
	return result
}

func buildWorkspaceCandidates(state, target Tree) []candidate {
	var result []candidate
	current := workspaceAllocation(state.Workspace)
	wanted := workspaceAllocation(target.Workspace)
	if state.ChargeType == "POST" {
		if current.Limit != wanted.Limit {
			result = append(result, newCandidate(ModifyWorkspacePostpaid, Ref{}, current, wanted, 0))
		}
		return result
	}
	if current.FixedCU != wanted.FixedCU || current.CrossZoneFixedCU != wanted.CrossZoneFixedCU {
		to := wanted
		to.Limit = to.TotalFixed() + current.AsCapacity().Elastic()
		result = append(result, newCandidate(ModifyWorkspaceFixed, Ref{}, current, to, 0))
	}
	currentElastic := current.AsCapacity().Elastic()
	wantedElastic := wanted.AsCapacity().Elastic()
	if currentElastic != wantedElastic {
		to := current
		to.Limit = to.TotalFixed() + wantedElastic
		action := ModifyWorkspaceElastic
		if currentElastic == 0 {
			action = EnableWorkspaceElastic
		}
		result = append(result, newCandidate(action, Ref{}, current, to, 0))
	}
	return result
}

func selectLegacyCandidate(state Tree, candidates []candidate, maxEndpointLimit CU, expansion bool) *candidate {
	selectedIndex := -1
	for i := range candidates {
		candidate := candidates[i]
		if candidate.expansion != expansion || !legacyCandidateIsSafe(state, candidate, maxEndpointLimit) {
			continue
		}
		if selectedIndex < 0 || betterLevel(candidate.level, candidates[selectedIndex].level, expansion) {
			selectedIndex = i
		}
	}
	if selectedIndex < 0 {
		return nil
	}
	selected := candidates[selectedIndex]
	return &selected
}

func legacyCandidateIsSafe(state Tree, candidate candidate, maxEndpointLimit CU) bool {
	if candidate.level == 0 && candidate.step.To.Limit > maxEndpointLimit {
		return false
	}
	next := cloneTree(state)
	applyCandidate(&next, candidate)
	_, err := Resolve(next)
	return err == nil
}

func legacySameCapacityTree(actual, desired Tree) bool {
	if workspaceAllocation(actual.Workspace) != workspaceAllocation(desired.Workspace) {
		return false
	}
	for _, wanted := range desired.Namespaces {
		observed := namespaceByName(actual, wanted.Name)
		if *observed.Capacity != *wanted.Capacity {
			return false
		}
		for _, wantedQueue := range wanted.Queues {
			observedQueue := queueByName(*observed, wantedQueue.Name)
			if *observedQueue.Capacity != *wantedQueue.Capacity {
				return false
			}
		}
	}
	return true
}

// PlanAuthoritative returns a deterministic forward-only sequence for a
// Workspace-owned namespace topology. The caller must read and replan after
// every step because the control plane is asynchronous.
func PlanAuthoritative(actual, desired Tree) ([]Step, error) {
	planning := newPlannerState(actual)
	return planWithState(actual, desired, planning)
}

type plannerState struct {
	temporaryGranted  bool
	createdNamespaces map[string]struct{}
}

func newPlannerState(actual Tree) *plannerState {
	fixedHeadroom, limitHeadroom := workspaceHeadroom(actual)
	return &plannerState{
		temporaryGranted:  fixedHeadroom >= minimumNamespaceCU && limitHeadroom >= minimumNamespaceCU,
		createdNamespaces: make(map[string]struct{}),
	}
}

func (s *plannerState) clone() *plannerState {
	result := &plannerState{
		temporaryGranted:  s.temporaryGranted,
		createdNamespaces: make(map[string]struct{}, len(s.createdNamespaces)),
	}
	for name := range s.createdNamespaces {
		result.createdNamespaces[name] = struct{}{}
	}
	return result
}

func (s *plannerState) noteCompleted(step Step) {
	if step.Temporary {
		s.temporaryGranted = true
	}
	if step.Action == CreateNamespace {
		s.createdNamespaces[step.Ref.Namespace] = struct{}{}
	}
}

func planWithState(actual, desired Tree, prior *plannerState) ([]Step, error) {
	if err := validateAuthoritativeChargeTypes(actual, desired); err != nil {
		return nil, err
	}
	if actual.ChargeType != desired.ChargeType {
		return nil, fmt.Errorf("workspace charge type cannot change from %q to %q", actual.ChargeType, desired.ChargeType)
	}
	if actual.Workspace.HA != desired.Workspace.HA {
		return nil, fmt.Errorf("workspace high availability mode cannot change in place")
	}
	if err := ValidateDesired(desired); err != nil {
		return nil, fmt.Errorf("desired capacity tree: %w", err)
	}
	if err := validateRetainedNamespaceTypes(actual, desired); err != nil {
		return nil, err
	}
	if err := validateObserved(actual); err != nil {
		return nil, fmt.Errorf("actual capacity tree: %w", err)
	}
	if err := validateRetainedNamespaces(actual, desired); err != nil {
		return nil, err
	}
	if err := validateUsedCapacity(actual, desired); err != nil {
		return nil, err
	}
	actualElastic := actual.Workspace.AsCapacity().Elastic()
	desiredElastic := desired.Workspace.AsCapacity().Elastic()
	if actual.ChargeType == "PRE" && actualElastic > 0 && desiredElastic == 0 {
		return nil, fmt.Errorf("cannot reduce workspace elastic CU to zero through the supported public API; disable it manually before retrying")
	}

	state := cloneTree(actual)
	target := cloneTree(desired)
	planning := prior.clone()
	maxEndpointLimit := actual.Workspace.Limit
	if target.Workspace.Limit > maxEndpointLimit {
		maxEndpointLimit = target.Workspace.Limit
	}
	var steps []Step
	for attempts := 0; !sameCapacityTree(state, target); attempts++ {
		if attempts > (len(state.Namespaces)+len(target.Namespaces)+1)*8+8 {
			return nil, fmt.Errorf("capacity planner did not converge")
		}

		candidates := buildCandidates(state, target, planning)
		selected := selectCandidate(state, candidates, maxEndpointLimit, true)
		if selected == nil {
			selected = selectCandidate(state, candidates, maxEndpointLimit, false)
		}
		if selected == nil && hasMissingNamespace(state, target) {
			selected = temporaryWorkspaceCandidate(state, planning)
		}
		if selected == nil {
			return nil, fmt.Errorf("no safe capacity transition exists without exceeding endpoint capacity or violating child allocations")
		}

		applyCandidate(&state, *selected)
		steps = append(steps, selected.step)
		planning.noteCompleted(selected.step)
		if selected.step.Action == DeleteNamespace {
			return steps, nil
		}
	}
	return steps, nil
}

func validateAuthoritativeChargeTypes(actual, desired Tree) error {
	if actual.ChargeType != "PRE" {
		return fmt.Errorf("authoritative capacity planning supports only PRE actual workspace charge type, got %q", actual.ChargeType)
	}
	if desired.ChargeType != "PRE" {
		return fmt.Errorf("authoritative capacity planning supports only PRE desired workspace charge type, got %q", desired.ChargeType)
	}
	return nil
}

func validateWorkspace(chargeType string, workspace WorkspaceCapacity) error {
	capacity := workspace.AsCapacity()
	if err := capacity.Validate(); err != nil {
		return fmt.Errorf("workspace capacity: %w", err)
	}
	if err := validateUsedCU("workspace", workspace.Used, capacity.Limit); err != nil {
		return err
	}
	if workspace.HA {
		if workspace.CrossZoneFixedCU <= 0 {
			return fmt.Errorf("HA workspace cross-zone fixed CU must be greater than zero")
		}
	} else if workspace.CrossZoneFixedCU != 0 {
		return fmt.Errorf("non-HA workspace cannot configure cross-zone fixed CU")
	}
	switch chargeType {
	case "PRE":
		if workspace.TotalFixed() <= 0 {
			return fmt.Errorf("PRE workspace fixed CU total must be greater than zero")
		}
	case "POST":
		if workspace.HA {
			return fmt.Errorf("POST workspace cannot use high availability")
		}
		if workspace.TotalFixed() != 0 {
			return fmt.Errorf("POST workspace fixed CU must be zero")
		}
		if workspace.Limit <= 0 {
			return fmt.Errorf("POST workspace CU limit must be greater than zero")
		}
	default:
		return fmt.Errorf("unsupported charge type %q", chargeType)
	}
	return nil
}

func validateAuthoritativeWorkspace(workspace WorkspaceCapacity) error {
	if err := validateAuthoritativeWorkspaceIntegers(workspace); err != nil {
		return err
	}
	return validateAuthoritativeWorkspaceMode(workspace)
}

func validateAuthoritativeWorkspaceIntegers(workspace WorkspaceCapacity) error {
	if err := validateIntegerCU("workspace fixed CU", workspace.FixedCU); err != nil {
		return err
	}
	if err := validateIntegerCU("workspace cross-zone fixed CU", workspace.CrossZoneFixedCU); err != nil {
		return err
	}
	if err := validateIntegerCU("workspace max CU", workspace.Limit); err != nil {
		return err
	}
	return nil
}

func validateAuthoritativeWorkspaceMode(workspace WorkspaceCapacity) error {
	if workspace.HA {
		if workspace.FixedCU != 0 || workspace.CrossZoneFixedCU <= 0 {
			return fmt.Errorf("authoritative HA workspace must use the pure cross-zone fixed CU pool")
		}
		return nil
	}
	if workspace.FixedCU <= 0 || workspace.CrossZoneFixedCU != 0 {
		return fmt.Errorf("authoritative non-HA workspace must use the pure single-zone fixed CU pool")
	}
	return nil
}

func validateAuthoritativeNamespace(namespace Namespace) error {
	if err := validateIntegerCU(fmt.Sprintf("namespace %q fixed CU", namespace.Name), namespace.Capacity.Fixed); err != nil {
		return err
	}
	if err := validateIntegerCU(fmt.Sprintf("namespace %q max CU", namespace.Name), namespace.Capacity.Limit); err != nil {
		return err
	}
	return nil
}

func validateIntegerCU(name string, value CU) error {
	if value%2 != 0 {
		return fmt.Errorf("%s must be an integer CU, got %v", name, value.Float64())
	}
	return nil
}

func validateObserved(tree Tree) error {
	if err := validateWorkspace(tree.ChargeType, tree.Workspace); err != nil {
		return err
	}
	if err := validateAuthoritativeWorkspace(tree.Workspace); err != nil {
		return err
	}
	seenNamespaces := make(map[string]struct{}, len(tree.Namespaces))
	for _, namespace := range tree.Namespaces {
		if namespace.Name == "" {
			return fmt.Errorf("namespace name must not be empty")
		}
		if _, exists := seenNamespaces[namespace.Name]; exists {
			return fmt.Errorf("workspace has duplicate namespace %q", namespace.Name)
		}
		seenNamespaces[namespace.Name] = struct{}{}
		if namespace.Capacity == nil {
			return fmt.Errorf("namespace %q capacity is missing", namespace.Name)
		}
		if err := namespace.Capacity.Validate(); err != nil {
			return fmt.Errorf("namespace %q: %w", namespace.Name, err)
		}
		if err := validateAuthoritativeNamespace(namespace); err != nil {
			return err
		}
		if namespace.CrossZone != tree.Workspace.HA {
			return fmt.Errorf("namespace %q cross-zone type does not match workspace fixed CU pool", namespace.Name)
		}
		seenQueues := make(map[string]struct{}, len(namespace.Queues))
		for _, queue := range namespace.Queues {
			if queue.Name == "" {
				return fmt.Errorf("namespace %q has queue with empty name", namespace.Name)
			}
			if _, exists := seenQueues[queue.Name]; exists {
				return fmt.Errorf("namespace %q has duplicate queue %q", namespace.Name, queue.Name)
			}
			seenQueues[queue.Name] = struct{}{}
			if queue.Capacity == nil {
				return fmt.Errorf("queue %q/%q capacity is missing", namespace.Name, queue.Name)
			}
			if err := queue.Capacity.Validate(); err != nil {
				return fmt.Errorf("queue %q/%q: %w", namespace.Name, queue.Name, err)
			}
		}
	}
	return nil
}

func validateRetainedNamespaces(actual, desired Tree) error {
	for _, wanted := range desired.Namespaces {
		observed := namespaceByName(actual, wanted.Name)
		if observed == nil {
			continue
		}
		if queueByName(*observed, "default-queue") == nil {
			return &TopologyError{message: fmt.Sprintf("declared queue %q/default-queue does not exist in the workspace", wanted.Name), retryable: true}
		}
		for _, queue := range observed.Queues {
			if queue.Name != "default-queue" {
				capacity := *queue.Capacity
				return &TopologyError{message: fmt.Sprintf(
					"retained namespace %q contains unsupported queue %q (fixed_cu=%v, max_cu_limit=%v, used_cu=%v)",
					wanted.Name, queue.Name, capacity.Fixed.Float64(), capacity.Limit.Float64(), queue.Used,
				)}
			}
		}
	}
	return nil
}

func validateRetainedNamespaceTypes(actual, desired Tree) error {
	for _, wanted := range desired.Namespaces {
		observed := namespaceByName(actual, wanted.Name)
		if observed != nil && observed.CrossZone != wanted.CrossZone {
			return fmt.Errorf("retained namespace %q type mismatch: observed cross-zone=%t, desired cross-zone=%t", wanted.Name, observed.CrossZone, wanted.CrossZone)
		}
	}
	return nil
}

func validateUsedCapacity(actual, desired Tree) error {
	if !hasUndeclaredNamespace(actual, desired) && actual.Workspace.Used > desired.Workspace.Limit.Float64() {
		return fmt.Errorf("workspace desired limit %v is below used CU %v", desired.Workspace.Limit.Float64(), actual.Workspace.Used)
	}
	for _, wanted := range desired.Namespaces {
		observed := namespaceByName(actual, wanted.Name)
		if observed == nil {
			continue
		}
		if observed.Used > wanted.Capacity.Limit.Float64() {
			return fmt.Errorf("namespace %q desired limit %v is below used CU %v", wanted.Name, wanted.Capacity.Limit.Float64(), observed.Used)
		}
		observedQueue := queueByName(*observed, "default-queue")
		if observedQueue != nil && observedQueue.Used > wanted.Capacity.Limit.Float64() {
			return fmt.Errorf("queue %q/default-queue desired limit %v is below used CU %v", wanted.Name, wanted.Capacity.Limit.Float64(), observedQueue.Used)
		}
	}
	return nil
}

func hasUndeclaredNamespace(actual, desired Tree) bool {
	for _, observed := range actual.Namespaces {
		if namespaceByName(desired, observed.Name) == nil {
			return true
		}
	}
	return false
}

func buildCandidates(state, target Tree, planning *plannerState) []candidate {
	result := make([]candidate, 0, capacityNodeCount(state)+capacityNodeCount(target))
	currentWorkspace := workspaceAllocation(state.Workspace)
	targetWorkspace := workspaceAllocation(target.Workspace)
	if state.ChargeType == "POST" {
		if currentWorkspace.Limit != targetWorkspace.Limit {
			result = append(result, newCandidate(ModifyWorkspacePostpaid, Ref{}, currentWorkspace, targetWorkspace, 0))
		}
	} else {
		if currentWorkspace.FixedCU != targetWorkspace.FixedCU || currentWorkspace.CrossZoneFixedCU != targetWorkspace.CrossZoneFixedCU {
			to := targetWorkspace
			to.Limit = to.TotalFixed() + currentWorkspace.AsCapacity().Elastic()
			result = append(result, newCandidate(ModifyWorkspaceFixed, Ref{}, currentWorkspace, to, 0))
		}
		currentElastic := currentWorkspace.AsCapacity().Elastic()
		targetElastic := targetWorkspace.AsCapacity().Elastic()
		if currentElastic != targetElastic {
			to := currentWorkspace
			to.Limit = to.TotalFixed() + targetElastic
			action := ModifyWorkspaceElastic
			if currentElastic == 0 {
				action = EnableWorkspaceElastic
			}
			result = append(result, newCandidate(action, Ref{}, currentWorkspace, to, 0))
		}
	}

	missingNamespace := false
	for _, wanted := range sortedNamespaces(target.Namespaces) {
		observed := namespaceByName(state, wanted.Name)
		if observed == nil {
			missingNamespace = true
			if state.ChargeType == "POST" {
				continue
			}
			minimum := Allocation{FixedCU: minimumNamespaceCU, Limit: minimumNamespaceCU}
			candidate := newCandidate(CreateNamespace, Ref{Namespace: wanted.Name}, Allocation{}, minimum, 1)
			candidate.step.NamespaceCrossZone = wanted.CrossZone
			result = append(result, candidate)
			continue
		}
		from := capacityAllocation(*observed.Capacity)
		to := capacityAllocation(*wanted.Capacity)
		result = appendCapacityDimensionCandidates(result, ModifyNamespace, Ref{Namespace: wanted.Name}, from, to, 2)
		observedQueue := queueByName(*observed, "default-queue")
		from = capacityAllocation(*observedQueue.Capacity)
		to = capacityAllocation(*wanted.Queues[0].Capacity)
		result = appendCapacityDimensionCandidates(result, ModifyQueue, Ref{Namespace: wanted.Name, Queue: "default-queue"}, from, to, 3)
	}

	// Missing namespaces reserve available headroom before retained allocations
	// expand. A true surplus (more actual than desired namespaces) may be deleted
	// first to avoid temporary capacity; paired rename sources remain until a
	// replacement has been created.
	if !missingNamespace || len(planning.createdNamespaces) > 0 || len(state.Namespaces) > len(target.Namespaces) {
		for _, observed := range sortedNamespaces(state.Namespaces) {
			if namespaceByName(target, observed.Name) != nil || len(state.Namespaces) <= 1 {
				continue
			}
			result = append(result, newCandidate(DeleteNamespace, Ref{Namespace: observed.Name}, capacityAllocation(*observed.Capacity), Allocation{}, 4))
		}
	}
	return result
}

func appendCapacityDimensionCandidates(result []candidate, action Action, ref Ref, from, target Allocation, level int) []candidate {
	if from.FixedCU != target.FixedCU || from.CrossZoneFixedCU != target.CrossZoneFixedCU {
		to := from
		to.FixedCU = target.FixedCU
		to.CrossZoneFixedCU = target.CrossZoneFixedCU
		if err := to.AsCapacity().Validate(); err == nil {
			result = append(result, newCandidate(action, ref, from, to, level))
		}
	}
	if from.Limit != target.Limit {
		to := from
		to.Limit = target.Limit
		if err := to.AsCapacity().Validate(); err == nil {
			result = append(result, newCandidate(action, ref, from, to, level))
		}
	}
	return result
}

func sortedNamespaces(input []Namespace) []Namespace {
	result := append([]Namespace(nil), input...)
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func hasMissingNamespace(state, target Tree) bool {
	for _, wanted := range target.Namespaces {
		if namespaceByName(state, wanted.Name) == nil {
			return true
		}
	}
	return false
}

func temporaryWorkspaceCandidate(state Tree, planning *plannerState) *candidate {
	if state.ChargeType != "PRE" {
		return nil
	}
	if planning.temporaryGranted {
		return nil
	}
	current := workspaceAllocation(state.Workspace)
	fixedHeadroom, limitHeadroom := workspaceHeadroom(state)
	if fixedHeadroom >= minimumNamespaceCU && limitHeadroom >= minimumNamespaceCU {
		return nil
	}
	to := current
	if state.Workspace.HA {
		to.CrossZoneFixedCU += minimumNamespaceCU
	} else {
		to.FixedCU += minimumNamespaceCU
	}
	to.Limit += minimumNamespaceCU
	return &candidate{
		step:      Step{Action: ModifyWorkspaceFixed, From: current, To: to, Temporary: true},
		level:     0,
		expansion: true,
	}
}

func workspaceHeadroom(tree Tree) (CU, CU) {
	workspace := tree.Workspace
	var allocatedFixed, allocatedLimit CU
	for _, namespace := range tree.Namespaces {
		if namespace.Capacity == nil {
			continue
		}
		if namespace.CrossZone == workspace.HA {
			allocatedFixed += namespace.Capacity.Fixed
		}
		allocatedLimit += namespace.Capacity.Limit
	}
	poolFixed := workspace.FixedCU
	if workspace.HA {
		poolFixed = workspace.CrossZoneFixedCU
	}
	return poolFixed - allocatedFixed, workspace.Limit - allocatedLimit
}

func newCandidate(action Action, ref Ref, from, to Allocation, level int) candidate {
	return candidate{
		step:      Step{Action: action, Ref: ref, From: from, To: to},
		level:     level,
		expansion: to.TotalFixed() > from.TotalFixed() || to.Limit > from.Limit || action == CreateNamespace,
	}
}

func selectCandidate(state Tree, candidates []candidate, maxEndpointLimit CU, expansion bool) *candidate {
	selectedIndex := -1
	for i := range candidates {
		candidate := candidates[i]
		if candidate.expansion != expansion || !candidateIsSafe(state, candidate, maxEndpointLimit) {
			continue
		}
		if selectedIndex < 0 || betterLevel(candidate.level, candidates[selectedIndex].level, expansion) {
			selectedIndex = i
		}
	}
	if selectedIndex < 0 {
		return nil
	}
	selected := candidates[selectedIndex]
	return &selected
}

func betterLevel(candidate, selected int, expansion bool) bool {
	if expansion {
		return candidate < selected
	}
	return candidate > selected
}

func candidateIsSafe(state Tree, candidate candidate, maxEndpointLimit CU) bool {
	if candidate.level == 0 && candidate.step.To.Limit > maxEndpointLimit {
		return false
	}
	next := cloneTree(state)
	applyCandidate(&next, candidate)
	return validatePlanningState(next) == nil
}

func validatePlanningState(tree Tree) error {
	if err := validateObserved(tree); err != nil {
		return err
	}
	var fixed, crossZoneFixed, limit CU
	for _, namespace := range tree.Namespaces {
		if namespace.CrossZone {
			crossZoneFixed += namespace.Capacity.Fixed
		} else {
			fixed += namespace.Capacity.Fixed
		}
		limit += namespace.Capacity.Limit
		defaultQueue := queueByName(namespace, "default-queue")
		if defaultQueue == nil {
			continue // Undeclared namespaces may retain arbitrary cloud topology until deletion.
		}
		if defaultQueue.Capacity.Fixed > namespace.Capacity.Fixed || defaultQueue.Capacity.Limit > namespace.Capacity.Limit {
			return fmt.Errorf("queue %q/default-queue allocation exceeds namespace allocation", namespace.Name)
		}
	}
	if fixed > tree.Workspace.FixedCU || crossZoneFixed > tree.Workspace.CrossZoneFixedCU || limit > tree.Workspace.Limit {
		return fmt.Errorf("namespace allocations exceed workspace capacity")
	}
	return nil
}

func applyCandidate(state *Tree, candidate candidate) {
	to := candidate.step.To
	switch candidate.step.Action {
	case ModifyWorkspaceFixed, ModifyWorkspacePostpaid, EnableWorkspaceElastic, ModifyWorkspaceElastic:
		state.Workspace.FixedCU = to.FixedCU
		state.Workspace.CrossZoneFixedCU = to.CrossZoneFixedCU
		state.Workspace.Limit = to.Limit
	case CreateNamespace:
		capacity := to.AsCapacity()
		state.Namespaces = append(state.Namespaces, Namespace{
			Name:      candidate.step.Ref.Namespace,
			CrossZone: candidate.step.NamespaceCrossZone,
			Capacity:  &capacity,
			Queues:    []Queue{{Name: "default-queue", Capacity: cloneCapacity(&capacity)}},
		})
	case DeleteNamespace:
		for i := range state.Namespaces {
			if state.Namespaces[i].Name == candidate.step.Ref.Namespace {
				state.Namespaces = append(state.Namespaces[:i], state.Namespaces[i+1:]...)
				return
			}
		}
	case ModifyNamespace:
		namespace := namespaceByName(*state, candidate.step.Ref.Namespace)
		capacity := to.AsCapacity()
		namespace.Capacity = &capacity
	case ModifyQueue:
		namespace := namespaceByName(*state, candidate.step.Ref.Namespace)
		queue := queueByName(*namespace, candidate.step.Ref.Queue)
		capacity := to.AsCapacity()
		queue.Capacity = &capacity
	}
}

func sameCapacityTree(actual, desired Tree) bool {
	if workspaceAllocation(actual.Workspace) != workspaceAllocation(desired.Workspace) || len(actual.Namespaces) != len(desired.Namespaces) {
		return false
	}
	for _, wanted := range desired.Namespaces {
		observed := namespaceByName(actual, wanted.Name)
		if observed == nil || observed.CrossZone != wanted.CrossZone || observed.Capacity == nil || *observed.Capacity != *wanted.Capacity || len(observed.Queues) != 1 {
			return false
		}
		queue := queueByName(*observed, "default-queue")
		if queue == nil || queue.Capacity == nil || *queue.Capacity != *wanted.Capacity {
			return false
		}
	}
	return true
}

func namespaceByName(tree Tree, name string) *Namespace {
	for i := range tree.Namespaces {
		if tree.Namespaces[i].Name == name {
			return &tree.Namespaces[i]
		}
	}
	return nil
}

func queueByName(namespace Namespace, name string) *Queue {
	for i := range namespace.Queues {
		if namespace.Queues[i].Name == name {
			return &namespace.Queues[i]
		}
	}
	return nil
}

func workspaceAllocation(capacity WorkspaceCapacity) Allocation {
	return Allocation{
		FixedCU:          capacity.FixedCU,
		CrossZoneFixedCU: capacity.CrossZoneFixedCU,
		Limit:            capacity.Limit,
	}
}

func capacityAllocation(capacity Capacity) Allocation {
	return Allocation{FixedCU: capacity.Fixed, Limit: capacity.Limit}
}

func capacityNodeCount(tree Tree) int {
	count := 1 + len(tree.Namespaces)
	for _, namespace := range tree.Namespaces {
		count += len(namespace.Queues)
	}
	return count
}
