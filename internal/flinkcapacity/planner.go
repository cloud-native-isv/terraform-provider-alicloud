package flinkcapacity

import (
	"context"
	"fmt"
	"math"
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

const (
	minimumNamespaceCU CU = 2 // one CU, represented in half-CU units.
	// FOAS serializes CPU and MemoryGB through int32 fields. Authoritative
	// capacity is integral CU, so floor(MaxInt32/4) external CU is the largest
	// component whose MemoryGB=CU*4 remains representable. Keep the multiply
	// after the division so the half-CU constant is overflow-safe.
	maximumFlinkSerializableComponentCU CU = CU(((1 << 31) - 1) / 4 * 2)
)

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
	return validateDesiredContext(context.Background(), tree)
}

func validateDesiredContext(ctx context.Context, tree Tree) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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
		if err := ctx.Err(); err != nil {
			return err
		}
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
		if err := ctx.Err(); err != nil {
			return err
		}
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
	return PlanAuthoritativeContext(context.Background(), actual, desired)
}

// PlanAuthoritativeContext is the production authoritative planner entrypoint.
// It is deliberately stateless across calls: every successful cloud write is
// followed by a complete Read and the next plan is derived from that observed
// tree. A returned tail is only a feasibility witness; callers must apply the
// first step and replan.
func PlanAuthoritativeContext(ctx context.Context, actual, desired Tree) ([]Step, error) {
	steps, _, err := planAuthoritativeWithLimits(ctx, actual, desired, plannerLimits{})
	return steps, err
}

const (
	defaultMaxPlanSteps        = 4096
	defaultMaxPlannerWorkUnits = 1_000_000
)

type plannerLimits struct {
	MaxSteps     int
	MaxWorkUnits int
}

type plannerStats struct {
	Steps     int
	WorkUnits int
}

// PlannerBudgetError is distinct from dependency stall. Budget exhaustion is
// fail-closed and must never authorize the one-CU fallback phase.
type PlannerBudgetError struct {
	Kind           string
	Phase          string
	Steps          int
	WorkUnits      int
	NodeCount      int
	EstimatedBound int
}

func (e *PlannerBudgetError) Error() string {
	return fmt.Sprintf("authoritative capacity planner %s budget exceeded in %s phase (steps=%d, work_units=%d, nodes=%d, estimated_step_bound=%d)", e.Kind, e.Phase, e.Steps, e.WorkUnits, e.NodeCount, e.EstimatedBound)
}

type plannerBudget struct {
	ctx            context.Context
	limits         plannerLimits
	stats          plannerStats
	nodeCount      int
	preflightWork  int
	estimatedBound int
}

func newPlannerBudget(ctx context.Context, actual, desired Tree, limits plannerLimits) (*plannerBudget, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limits.MaxSteps <= 0 {
		limits.MaxSteps = defaultMaxPlanSteps
	}
	if limits.MaxWorkUnits <= 0 {
		limits.MaxWorkUnits = defaultMaxPlannerWorkUnits
	}
	actualNodes, err := contextNodeCount(ctx, actual)
	if err != nil {
		return nil, err
	}
	desiredNodes, err := contextNodeCount(ctx, desired)
	if err != nil {
		return nil, err
	}
	nodes := actualNodes
	if desiredNodes > nodes {
		nodes = desiredNodes
	}
	return &plannerBudget{
		ctx:            ctx,
		limits:         limits,
		nodeCount:      nodes,
		preflightWork:  saturatingAddInt(actualNodes, desiredNodes),
		estimatedBound: estimateAuthoritativePlanSteps(actual, desired),
	}, nil
}

func contextNodeCount(ctx context.Context, tree Tree) (int, error) {
	count := 1
	for i := range tree.Namespaces {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		count = saturatingAddInt(count, 1, len(tree.Namespaces[i].Queues))
	}
	return count, ctx.Err()
}

func estimateAuthoritativePlanSteps(actual, desired Tree) int {
	fixedSwapUnits := absoluteCUDifference(
		activeWorkspaceFixed(actual.Workspace),
		activeWorkspaceFixed(desired.Workspace),
	) / uint64(minimumNamespaceCU)
	return saturatingAddInt(
		8,
		saturatingMultiplyInt(2, saturatingUint64ToInt(fixedSwapUnits)),
		saturatingMultiplyInt(9, len(desired.Namespaces)),
		len(actual.Namespaces),
	)
}

func absoluteCUDifference(first, second CU) uint64 {
	// Unsigned subtraction is defined modulo 2^64, which preserves the exact
	// distance even when the signed operands straddle MinInt64 and MaxInt64.
	if first >= second {
		return uint64(first) - uint64(second)
	}
	return uint64(second) - uint64(first)
}

func saturatingUint64ToInt(value uint64) int {
	maximum := maximumInt()
	if value > uint64(maximum) {
		return maximum
	}
	return int(value)
}

func saturatingMultiplyInt(first, second int) int {
	maximum := maximumInt()
	if first < 0 || second < 0 {
		return maximum
	}
	if first != 0 && second > maximum/first {
		return maximum
	}
	return first * second
}

func saturatingAddInt(values ...int) int {
	result := 0
	for _, value := range values {
		next, overflow := addNonNegativeInt(result, value)
		if overflow {
			return maximumInt()
		}
		result = next
	}
	return result
}

func addNonNegativeInt(current, increment int) (int, bool) {
	maximum := maximumInt()
	if current < 0 || increment < 0 || current > maximum-increment {
		return maximum, true
	}
	return current + increment, false
}

func maximumInt() int {
	return int(^uint(0) >> 1)
}

// minimumAuthoritativePlanSteps is deliberately conservative. It counts only
// mutations that distinct API component boundaries make unavoidable. When a
// Delete frontier is present it stops at that frontier rather than counting
// unknown post-Read work. Exceeding MaxSteps here is therefore a proof that no
// witness can fit, not a heuristic estimate.
func minimumAuthoritativePlanSteps(ctx context.Context, actual, desired Tree) (int, error) {
	actualByName, err := indexNamespacesContext(ctx, actual)
	if err != nil {
		return 0, err
	}
	desiredByName, err := indexNamespacesContext(ctx, desired)
	if err != nil {
		return 0, err
	}

	missing, undeclared := 0, 0
	for name := range desiredByName {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if actualByName[name] == nil {
			missing = saturatingAddInt(missing, 1)
		}
	}
	for name := range actualByName {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if desiredByName[name] == nil {
			undeclared = saturatingAddInt(undeclared, 1)
		}
	}
	if undeclared > missing {
		return 1, nil // A strict surplus reaches a Delete frontier immediately.
	}
	if missing > 0 {
		minimum := 1 // at least one Create is unavoidable.
		if undeclared > 0 {
			minimum = saturatingAddInt(minimum, 1) // paired Delete frontier.
		}
		return minimum, nil
	}
	if undeclared > 0 {
		return 1, nil
	}

	steps := 0
	if activeWorkspaceFixed(actual.Workspace) != activeWorkspaceFixed(desired.Workspace) {
		steps = saturatingAddInt(steps, 1)
	}
	if actual.Workspace.AsCapacity().Elastic() != desired.Workspace.AsCapacity().Elastic() {
		steps = saturatingAddInt(steps, 1)
	}
	for _, wanted := range desired.Namespaces {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		observed := actualByName[wanted.Name]
		if observed.Capacity.Fixed != wanted.Capacity.Fixed || observed.Capacity.Limit != wanted.Capacity.Limit {
			steps = saturatingAddInt(steps, 1)
		}
		observedQueue := queueByName(*observed, "default-queue")
		wantedQueue := queueByName(wanted, "default-queue")
		if observedQueue.Capacity.Fixed != wantedQueue.Capacity.Fixed || observedQueue.Capacity.Limit != wantedQueue.Capacity.Limit {
			steps = saturatingAddInt(steps, 1)
		}
	}
	return steps, ctx.Err()
}

func (b *plannerBudget) checkContext() error {
	return b.ctx.Err()
}

func (b *plannerBudget) addWork(phase string, units int) error {
	if err := b.checkContext(); err != nil {
		return err
	}
	next, overflow := addNonNegativeInt(b.stats.WorkUnits, units)
	b.stats.WorkUnits = next
	if overflow || b.stats.WorkUnits > b.limits.MaxWorkUnits {
		return b.error("work", phase)
	}
	return nil
}

func (b *plannerBudget) addStep(phase string) error {
	if err := b.checkContext(); err != nil {
		return err
	}
	next, overflow := addNonNegativeInt(b.stats.Steps, 1)
	b.stats.Steps = next
	if overflow || b.stats.Steps > b.limits.MaxSteps {
		return b.error("steps", phase)
	}
	return nil
}

func (b *plannerBudget) witnessCapacity() int {
	if b.estimatedBound <= 0 {
		return 0
	}
	if b.estimatedBound < b.limits.MaxSteps {
		return b.estimatedBound
	}
	return b.limits.MaxSteps
}

func (b *plannerBudget) error(kind, phase string) error {
	return &PlannerBudgetError{
		Kind:           kind,
		Phase:          phase,
		Steps:          b.stats.Steps,
		WorkUnits:      b.stats.WorkUnits,
		NodeCount:      b.nodeCount,
		EstimatedBound: b.estimatedBound,
	}
}

func planAuthoritativeWithLimits(ctx context.Context, actual, desired Tree, limits plannerLimits) ([]Step, plannerStats, error) {
	budget, err := newPlannerBudget(ctx, actual, desired, limits)
	if err != nil {
		return nil, plannerStats{}, err
	}
	if err := budget.checkContext(); err != nil {
		return nil, budget.stats, err
	}
	if err := budget.addWork("preflight", budget.preflightWork); err != nil {
		return nil, budget.stats, err
	}
	if err := validateAuthoritativeChargeTypes(actual, desired); err != nil {
		return nil, budget.stats, err
	}
	if actual.ChargeType != desired.ChargeType {
		return nil, budget.stats, fmt.Errorf("workspace charge type cannot change from %q to %q", actual.ChargeType, desired.ChargeType)
	}
	if actual.Workspace.HA != desired.Workspace.HA {
		return nil, budget.stats, fmt.Errorf("workspace high availability mode cannot change in place")
	}
	if err := validateDesiredContext(budget.ctx, desired); err != nil {
		return nil, budget.stats, fmt.Errorf("desired capacity tree: %w", err)
	}
	if err := validateRetainedNamespaceTypes(budget.ctx, actual, desired); err != nil {
		return nil, budget.stats, err
	}
	if err := validateObservedContext(budget.ctx, actual); err != nil {
		return nil, budget.stats, fmt.Errorf("actual capacity tree: %w", err)
	}
	if err := validateRetainedNamespaces(budget.ctx, actual, desired); err != nil {
		return nil, budget.stats, err
	}
	if err := validateUsedCapacity(budget.ctx, actual, desired); err != nil {
		return nil, budget.stats, err
	}
	actualElastic := actual.Workspace.AsCapacity().Elastic()
	desiredElastic := desired.Workspace.AsCapacity().Elastic()
	if actual.ChargeType == "PRE" && actualElastic > 0 && desiredElastic == 0 {
		return nil, budget.stats, fmt.Errorf("cannot reduce workspace elastic CU to zero through the supported public API; disable it manually before retrying")
	}
	if err := validatePlanningStateContext(budget.ctx, actual); err != nil {
		return nil, budget.stats, fmt.Errorf("actual capacity tree: %w", err)
	}
	if err := budget.checkContext(); err != nil {
		return nil, budget.stats, err
	}
	minimumSteps, err := minimumAuthoritativePlanSteps(budget.ctx, actual, desired)
	if err != nil {
		return nil, budget.stats, err
	}
	if minimumSteps > budget.limits.MaxSteps {
		return nil, budget.stats, budget.error("steps", "preflight")
	}

	endpoint := maxCU(actual.Workspace.Limit, desired.Workspace.Limit)
	result, err := scheduleAuthoritativePhase(cloneTree(actual), desired, endpoint, endpoint, "endpoint", false, budget)
	if err != nil {
		return nil, budget.stats, err
	}
	if result.status != phaseDependencyStall {
		return result.steps, budget.stats, nil
	}

	fallbackView, err := buildCanonicalPhaseState(&actual, &desired, "fallback", budget)
	if err != nil {
		return nil, budget.stats, err
	}
	missing := fallbackView.missingNames
	fallback := endpoint
	if len(missing) > 0 {
		// An observed limit above desired may be a buffer from this operation or
		// unrelated/imported capacity. The current tree cannot distinguish those
		// histories. Endpoint scheduling has already reused any visible headroom;
		// if it still stalled, growing from this unknown peak would permit C0+2 on
		// retry. Preserve rename sources and fail closed instead.
		if actual.Workspace.Limit > desired.Workspace.Limit {
			return nil, budget.stats, noSafeTransitionError()
		}
		if endpoint > CU(math.MaxInt64)-minimumNamespaceCU {
			return nil, budget.stats, fmt.Errorf("workspace capacity cannot be temporarily expanded without overflow")
		}
		fallback = endpoint + minimumNamespaceCU
	} else if desired.Workspace.Limit <= CU(math.MaxInt64)-minimumNamespaceCU {
		fallback = maxCU(endpoint, desired.Workspace.Limit+minimumNamespaceCU)
	}
	if fallback == endpoint {
		return nil, budget.stats, noSafeTransitionError()
	}
	result, err = scheduleAuthoritativePhase(cloneTree(actual), desired, fallback, endpoint, "fallback", len(missing) > 0, budget)
	if err != nil {
		return nil, budget.stats, err
	}
	if result.status == phaseDependencyStall {
		return nil, budget.stats, noSafeTransitionError()
	}
	return result.steps, budget.stats, nil
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
	return validateObservedContext(context.Background(), tree)
}

func validateObservedContext(ctx context.Context, tree Tree) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateWorkspace(tree.ChargeType, tree.Workspace); err != nil {
		return err
	}
	if err := validateAuthoritativeWorkspace(tree.Workspace); err != nil {
		return err
	}
	seenNamespaces := make(map[string]struct{}, len(tree.Namespaces))
	for _, namespace := range tree.Namespaces {
		if err := ctx.Err(); err != nil {
			return err
		}
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
			if err := ctx.Err(); err != nil {
				return err
			}
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

func validateRetainedNamespaces(ctx context.Context, actual, desired Tree) error {
	actualByName, err := indexNamespacesContext(ctx, actual)
	if err != nil {
		return err
	}
	for _, wanted := range desired.Namespaces {
		if err := ctx.Err(); err != nil {
			return err
		}
		observed := actualByName[wanted.Name]
		if observed == nil {
			continue
		}
		if queueByName(*observed, "default-queue") == nil {
			return &TopologyError{message: fmt.Sprintf("declared queue %q/default-queue does not exist in the workspace", wanted.Name), retryable: true}
		}
		for _, queue := range observed.Queues {
			if err := ctx.Err(); err != nil {
				return err
			}
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

func validateRetainedNamespaceTypes(ctx context.Context, actual, desired Tree) error {
	actualByName, err := indexNamespacesContext(ctx, actual)
	if err != nil {
		return err
	}
	for _, wanted := range desired.Namespaces {
		if err := ctx.Err(); err != nil {
			return err
		}
		observed := actualByName[wanted.Name]
		if observed != nil && observed.CrossZone != wanted.CrossZone {
			return fmt.Errorf("retained namespace %q type mismatch: observed cross-zone=%t, desired cross-zone=%t", wanted.Name, observed.CrossZone, wanted.CrossZone)
		}
	}
	return nil
}

func validateUsedCapacity(ctx context.Context, actual, desired Tree) error {
	hasUndeclared, err := hasUndeclaredNamespace(ctx, actual, desired)
	if err != nil {
		return err
	}
	if !hasUndeclared && actual.Workspace.Used > desired.Workspace.Limit.Float64() {
		return fmt.Errorf("workspace desired limit %v is below used CU %v", desired.Workspace.Limit.Float64(), actual.Workspace.Used)
	}
	actualByName, err := indexNamespacesContext(ctx, actual)
	if err != nil {
		return err
	}
	for _, wanted := range desired.Namespaces {
		if err := ctx.Err(); err != nil {
			return err
		}
		observed := actualByName[wanted.Name]
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

func hasUndeclaredNamespace(ctx context.Context, actual, desired Tree) (bool, error) {
	desiredByName, err := indexNamespacesContext(ctx, desired)
	if err != nil {
		return false, err
	}
	for _, observed := range actual.Namespaces {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if desiredByName[observed.Name] == nil {
			return true, nil
		}
	}
	return false, ctx.Err()
}

func indexNamespacesContext(ctx context.Context, tree Tree) (map[string]*Namespace, error) {
	result := make(map[string]*Namespace, len(tree.Namespaces))
	for i := range tree.Namespaces {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result[tree.Namespaces[i].Name] = &tree.Namespaces[i]
	}
	return result, ctx.Err()
}

type phaseStatus uint8

const (
	phaseComplete phaseStatus = iota
	phaseDeleteFrontier
	phaseDependencyStall
)

type phaseResult struct {
	steps  []Step
	status phaseStatus
}

type canonicalPhaseState struct {
	actualByName    map[string]*Namespace
	desiredByName   map[string]*Namespace
	desiredNames    []string
	missingNames    []string
	undeclaredNames []string
	allocatedFixed  CU
	allocatedLimit  CU
	sameAsDesired   bool
}

// buildCanonicalPhaseState charges each node scan to the hard work budget and
// builds name indexes once per scheduler iteration. All later selection in the
// iteration is O(N), so independent namespace operations are never explored as
// permutations.
func buildCanonicalPhaseState(state, target *Tree, phase string, budget *plannerBudget) (canonicalPhaseState, error) {
	view := canonicalPhaseState{
		actualByName:  make(map[string]*Namespace, len(state.Namespaces)),
		desiredByName: make(map[string]*Namespace, len(target.Namespaces)),
	}
	for i := range state.Namespaces {
		if err := budget.addWork(phase, saturatingAddInt(1, len(state.Namespaces[i].Queues))); err != nil {
			return canonicalPhaseState{}, err
		}
		namespace := &state.Namespaces[i]
		view.actualByName[namespace.Name] = namespace
		if namespace.Capacity != nil && namespace.CrossZone == state.Workspace.HA {
			view.allocatedFixed += namespace.Capacity.Fixed
			view.allocatedLimit += namespace.Capacity.Limit
		}
	}
	for i := range target.Namespaces {
		if err := budget.addWork(phase, saturatingAddInt(1, len(target.Namespaces[i].Queues))); err != nil {
			return canonicalPhaseState{}, err
		}
		namespace := &target.Namespaces[i]
		view.desiredByName[namespace.Name] = namespace
		view.desiredNames = append(view.desiredNames, namespace.Name)
		if _, ok := view.actualByName[namespace.Name]; !ok {
			view.missingNames = append(view.missingNames, namespace.Name)
		}
	}
	for name := range view.actualByName {
		if _, ok := view.desiredByName[name]; !ok {
			view.undeclaredNames = append(view.undeclaredNames, name)
		}
	}
	if err := budget.addWork(phase, len(view.actualByName)); err != nil {
		return canonicalPhaseState{}, err
	}
	sort.Strings(view.desiredNames)
	sort.Strings(view.missingNames)
	sort.Strings(view.undeclaredNames)
	if err := budget.checkContext(); err != nil {
		return canonicalPhaseState{}, err
	}

	view.sameAsDesired = workspaceAllocation(state.Workspace) == workspaceAllocation(target.Workspace) && len(state.Namespaces) == len(target.Namespaces)
	if view.sameAsDesired {
		for _, name := range view.desiredNames {
			if err := budget.addWork(phase, 1); err != nil {
				return canonicalPhaseState{}, err
			}
			observed := view.actualByName[name]
			wanted := view.desiredByName[name]
			if observed == nil || observed.CrossZone != wanted.CrossZone || observed.Capacity == nil || *observed.Capacity != *wanted.Capacity || len(observed.Queues) != 1 {
				view.sameAsDesired = false
				break
			}
			queue := queueByName(*observed, "default-queue")
			if queue == nil || queue.Capacity == nil || *queue.Capacity != *wanted.Capacity {
				view.sameAsDesired = false
				break
			}
		}
	}
	return view, nil
}

// scheduleAuthoritativePhase is intentionally a monotonic phase scheduler,
// not a graph search. Independent namespace mutations are ordered by stable
// name and never permuted. Target-directed child contractions run bottom-up;
// expansions run top-down. The only non-target mutation is one pure active
// fixed-pool topology buffer. Every Delete is an unconditional observation
// frontier, because Workspace used CU can only be refreshed by a real Read.
//
// Termination follows the lexicographic measure
// (missing+undeclared namespaces, create headroom deficit, target component
// distance). Create/Delete reduces the first term, the one topology buffer
// reduces the second, and every other step strictly reduces the third. Delete
// ends the invocation, so no simulated state crosses an observation frontier.
// Rebuilding canonical indexes costs O(N log N) per returned mutation and each
// safety check is O(N); total work is polynomial O(S*N log N), bounded by both
// max steps and charged work units. There is no recursion, visited graph, or
// enumeration of commutative namespace operation orders.
func scheduleAuthoritativePhase(state Tree, target Tree, ceiling, endpoint CU, phase string, topologyFallback bool, budget *plannerBudget) (phaseResult, error) {
	if err := budget.checkContext(); err != nil {
		return phaseResult{}, err
	}
	steps := make([]Step, 0, budget.witnessCapacity())
	for {
		if err := budget.checkContext(); err != nil {
			return phaseResult{}, err
		}
		view, err := buildCanonicalPhaseState(&state, &target, phase, budget)
		if err != nil {
			return phaseResult{}, err
		}
		if view.sameAsDesired {
			return phaseResult{steps: steps, status: phaseComplete}, nil
		}

		missing, undeclared := view.missingNames, view.undeclaredNames
		if len(undeclared) > len(missing) {
			step := deleteNamespaceStep(state, undeclared[0])
			if len(state.Namespaces) <= 1 {
				return phaseResult{status: phaseDependencyStall}, nil
			}
			if err := appendScheduledStep(&state, &steps, step, phase, budget); err != nil {
				return phaseResult{}, err
			}
			return phaseResult{steps: steps, status: phaseDeleteFrontier}, nil
		}

		if len(missing) > 0 {
			fixedHeadroom := activeWorkspaceFixed(state.Workspace) - view.allocatedFixed
			limitHeadroom := state.Workspace.Limit - view.allocatedLimit
			if fixedHeadroom >= minimumNamespaceCU && limitHeadroom >= minimumNamespaceCU {
				wanted := view.desiredByName[missing[0]]
				step := Step{
					Action:             CreateNamespace,
					Ref:                Ref{Namespace: missing[0]},
					To:                 Allocation{FixedCU: minimumNamespaceCU, Limit: minimumNamespaceCU},
					NamespaceCrossZone: wanted.CrossZone,
				}
				safe, err := authoritativeStepIsSafe(state, step, ceiling, phase, budget)
				if err != nil {
					return phaseResult{}, err
				}
				if !safe {
					return phaseResult{status: phaseDependencyStall}, nil
				}
				if err := appendScheduledStep(&state, &steps, step, phase, budget); err != nil {
					return phaseResult{}, err
				}
				continue
			}

			step, ok, err := nextTopologyPreparationStep(state, target, view, ceiling, endpoint, phase, topologyFallback, budget)
			if err != nil {
				return phaseResult{}, err
			}
			if ok {
				if err := appendScheduledStep(&state, &steps, step, phase, budget); err != nil {
					return phaseResult{}, err
				}
				continue
			}

			buffer, available := topologyBufferStep(state)
			if !available {
				return phaseResult{status: phaseDependencyStall}, nil
			}
			safe, err := authoritativeStepIsSafe(state, buffer, ceiling, phase, budget)
			if err != nil {
				return phaseResult{}, err
			}
			if !safe {
				return phaseResult{status: phaseDependencyStall}, nil
			}
			if err := appendScheduledStep(&state, &steps, buffer, phase, budget); err != nil {
				return phaseResult{}, err
			}
			continue
		}

		if len(undeclared) > 0 {
			step := deleteNamespaceStep(state, undeclared[0])
			if len(state.Namespaces) <= 1 {
				return phaseResult{status: phaseDependencyStall}, nil
			}
			if err := appendScheduledStep(&state, &steps, step, phase, budget); err != nil {
				return phaseResult{}, err
			}
			return phaseResult{steps: steps, status: phaseDeleteFrontier}, nil
		}

		step, ok, err := nextCapacityStep(state, target, view, ceiling, endpoint, phase, topologyFallback, budget)
		if err != nil {
			return phaseResult{}, err
		}
		if !ok {
			return phaseResult{status: phaseDependencyStall}, nil
		}
		if err := appendScheduledStep(&state, &steps, step, phase, budget); err != nil {
			return phaseResult{}, err
		}
	}
}

func nextTopologyPreparationStep(state, target Tree, view canonicalPhaseState, ceiling, endpoint CU, phase string, topologyFallback bool, budget *plannerBudget) (Step, bool, error) {
	if step, ok, err := nextChildContraction(state, view, ceiling, phase, budget); ok || err != nil {
		return step, ok, err
	}
	fixedHeadroom := activeWorkspaceFixed(state.Workspace) - view.allocatedFixed
	if fixedHeadroom < minimumNamespaceCU {
		if buffer, available := topologyBufferStep(state); available {
			safe, err := authoritativeStepIsSafe(state, buffer, ceiling, phase, budget)
			if err != nil || safe {
				return buffer, safe, err
			}
		}
	}
	step, ok, err := nextWorkspaceStep(state, target, view, ceiling, endpoint, phase, budget)
	if err != nil || !ok {
		return step, ok, err
	}
	// A missing-namespace fallback owns exactly one kind of endpoint-exceeding
	// mutation: topologyBufferStep. Target-directed Workspace steps may be
	// replayed while they remain within the endpoint, but an elastic or other
	// composition bridge must not consume the topology entitlement.
	if !workspaceStepAllowedInTopologyFallback(step, endpoint, topologyFallback) {
		return Step{}, false, nil
	}
	return step, true, nil
}

func nextCapacityStep(state, target Tree, view canonicalPhaseState, ceiling, endpoint CU, phase string, topologyFallback bool, budget *plannerBudget) (Step, bool, error) {
	if step, ok, err := nextChildContraction(state, view, ceiling, phase, budget); ok || err != nil {
		return step, ok, err
	}
	if step, ok, err := nextWorkspaceStep(state, target, view, ceiling, endpoint, phase, budget); err != nil {
		return Step{}, false, err
	} else if ok && workspaceStepAllowedInTopologyFallback(step, endpoint, topologyFallback) {
		return step, true, nil
	}
	return nextChildExpansion(state, view, ceiling, phase, budget)
}

func workspaceStepAllowedInTopologyFallback(step Step, endpoint CU, topologyFallback bool) bool {
	return !topologyFallback || step.To.Limit <= endpoint
}

func nextChildContraction(state Tree, view canonicalPhaseState, ceiling CU, phase string, budget *plannerBudget) (Step, bool, error) {
	for _, name := range view.desiredNames {
		if err := budget.addWork(phase, 1); err != nil {
			return Step{}, false, err
		}
		observed := view.actualByName[name]
		if observed == nil {
			continue
		}
		wanted := view.desiredByName[name]
		observedQueue := queueByName(*observed, "default-queue")
		wantedQueue := queueByName(*wanted, "default-queue")
		if observedQueue == nil || wantedQueue == nil {
			continue
		}
		if step, ok := contractionDimensionStep(ModifyQueue, Ref{Namespace: name, Queue: "default-queue"}, capacityAllocation(*observedQueue.Capacity), capacityAllocation(*wantedQueue.Capacity), observedQueue.Used); ok {
			safe, err := authoritativeStepIsSafe(state, step, ceiling, phase, budget)
			if err != nil || safe {
				return step, safe, err
			}
		}
	}
	for _, name := range view.desiredNames {
		if err := budget.addWork(phase, 1); err != nil {
			return Step{}, false, err
		}
		observed := view.actualByName[name]
		if observed == nil {
			continue
		}
		wanted := view.desiredByName[name]
		if step, ok := contractionDimensionStep(ModifyNamespace, Ref{Namespace: name}, capacityAllocation(*observed.Capacity), capacityAllocation(*wanted.Capacity), observed.Used); ok {
			safe, err := authoritativeStepIsSafe(state, step, ceiling, phase, budget)
			if err != nil || safe {
				return step, safe, err
			}
		}
	}
	return Step{}, false, nil
}

func contractionDimensionStep(action Action, ref Ref, from, target Allocation, used float64) (Step, bool) {
	to := from
	if from.TotalFixed() > target.TotalFixed() {
		to.FixedCU = target.FixedCU
		to.CrossZoneFixedCU = target.CrossZoneFixedCU
	}
	if from.Limit > target.Limit {
		minimum := maxCU(target.Limit, to.TotalFixed(), usedLimitCU(used))
		if minimum < from.Limit {
			to.Limit = minimum
		}
	}
	if to != from {
		return Step{Action: action, Ref: ref, From: from, To: to}, true
	}
	return Step{}, false
}

func nextWorkspaceStep(state, target Tree, view canonicalPhaseState, ceiling, endpoint CU, phase string, budget *plannerBudget) (Step, bool, error) {
	from := workspaceAllocation(state.Workspace)
	currentFixed := from.TotalFixed()
	currentElastic := from.AsCapacity().Elastic()
	targetAllocation := workspaceAllocation(target.Workspace)
	targetFixed := targetAllocation.TotalFixed()
	targetElastic := targetAllocation.AsCapacity().Elastic()
	allocatedFixed, allocatedLimit := view.allocatedFixed, view.allocatedLimit
	usedLimit := usedLimitCU(state.Workspace.Used)

	candidates := make([]Step, 0, 2)
	if currentFixed != targetFixed {
		newFixed := currentFixed
		if currentFixed < targetFixed {
			newFixed = minCU(targetFixed, ceiling-currentElastic)
		} else {
			newFixed = maxCU(targetFixed, allocatedFixed, allocatedLimit-currentElastic, usedLimit-currentElastic)
		}
		newFixed = clampNonNegative(newFixed)
		if movesToward(currentFixed, newFixed, targetFixed) {
			to := from
			setAllocationActiveFixed(&to, state.Workspace.HA, newFixed)
			to.Limit = newFixed + currentElastic
			candidates = append(candidates, Step{Action: ModifyWorkspaceFixed, From: from, To: to, Temporary: to.Limit > endpoint})
		}
	}
	if currentElastic != targetElastic {
		newElastic := currentElastic
		if currentElastic < targetElastic {
			newElastic = minCU(targetElastic, ceiling-currentFixed)
		} else {
			newElastic = maxCU(targetElastic, allocatedLimit-currentFixed, usedLimit-currentFixed, 0)
		}
		newElastic = clampNonNegative(newElastic)
		if movesToward(currentElastic, newElastic, targetElastic) {
			to := from
			to.Limit = currentFixed + newElastic
			action := ModifyWorkspaceElastic
			if currentElastic == 0 && newElastic > 0 {
				action = EnableWorkspaceElastic
			}
			candidates = append(candidates, Step{Action: action, From: from, To: to, Temporary: to.Limit > endpoint})
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		iDelta := candidates[i].To.Limit - candidates[i].From.Limit
		jDelta := candidates[j].To.Limit - candidates[j].From.Limit
		if (iDelta <= 0) != (jDelta <= 0) {
			return iDelta <= 0
		}
		return candidates[i].Action < candidates[j].Action
	})
	for _, step := range candidates {
		if err := budget.addWork(phase, 1); err != nil {
			return Step{}, false, err
		}
		safe, err := authoritativeStepIsSafe(state, step, ceiling, phase, budget)
		if err != nil {
			return Step{}, false, err
		}
		if safe {
			return step, true, nil
		}
	}
	return Step{}, false, nil
}

func nextChildExpansion(state Tree, view canonicalPhaseState, ceiling CU, phase string, budget *plannerBudget) (Step, bool, error) {
	for _, name := range view.desiredNames {
		if err := budget.addWork(phase, 1); err != nil {
			return Step{}, false, err
		}
		observed := view.actualByName[name]
		if observed == nil {
			continue
		}
		wanted := view.desiredByName[name]
		from := capacityAllocation(*observed.Capacity)
		targetAllocation := capacityAllocation(*wanted.Capacity)
		if step, ok := namespaceExpansionStep(state, view, name, from, targetAllocation); ok {
			safe, err := authoritativeStepIsSafe(state, step, ceiling, phase, budget)
			if err != nil || safe {
				return step, safe, err
			}
		}
	}
	for _, name := range view.desiredNames {
		if err := budget.addWork(phase, 1); err != nil {
			return Step{}, false, err
		}
		observed := view.actualByName[name]
		if observed == nil {
			continue
		}
		wanted := view.desiredByName[name]
		observedQueue := queueByName(*observed, "default-queue")
		wantedQueue := queueByName(*wanted, "default-queue")
		if observedQueue == nil || wantedQueue == nil {
			continue
		}
		if step, ok := queueExpansionStep(*observed, name, capacityAllocation(*observedQueue.Capacity), capacityAllocation(*wantedQueue.Capacity)); ok {
			safe, err := authoritativeStepIsSafe(state, step, ceiling, phase, budget)
			if err != nil || safe {
				return step, safe, err
			}
		}
	}
	return Step{}, false, nil
}

func namespaceExpansionStep(state Tree, view canonicalPhaseState, name string, from, target Allocation) (Step, bool) {
	otherFixed, otherLimit := view.allocatedFixed, view.allocatedLimit
	otherFixed -= from.TotalFixed()
	otherLimit -= from.Limit
	to := from
	if from.Limit < target.Limit {
		newLimit := minCU(target.Limit, state.Workspace.Limit-otherLimit)
		if newLimit > from.Limit {
			to.Limit = newLimit
		}
	}
	if from.TotalFixed() < target.TotalFixed() {
		newFixed := minCU(target.TotalFixed(), activeWorkspaceFixed(state.Workspace)-otherFixed, to.Limit)
		if newFixed > from.TotalFixed() {
			setAllocationActiveFixed(&to, false, newFixed)
		}
	}
	if to != from {
		return Step{Action: ModifyNamespace, Ref: Ref{Namespace: name}, From: from, To: to}, true
	}
	return Step{}, false
}

func queueExpansionStep(namespace Namespace, namespaceName string, from, target Allocation) (Step, bool) {
	to := from
	if from.Limit < target.Limit {
		newLimit := minCU(target.Limit, namespace.Capacity.Limit)
		if newLimit > from.Limit {
			to.Limit = newLimit
		}
	}
	if from.TotalFixed() < target.TotalFixed() {
		newFixed := minCU(target.TotalFixed(), namespace.Capacity.Fixed, to.Limit)
		if newFixed > from.TotalFixed() {
			setAllocationActiveFixed(&to, false, newFixed)
		}
	}
	if to != from {
		return Step{Action: ModifyQueue, Ref: Ref{Namespace: namespaceName, Queue: "default-queue"}, From: from, To: to}, true
	}
	return Step{}, false
}

func topologyBufferStep(state Tree) (Step, bool) {
	from := workspaceAllocation(state.Workspace)
	currentFixed := activeWorkspaceFixed(state.Workspace)
	if currentFixed < 0 || currentFixed > maximumFlinkSerializableComponentCU-minimumNamespaceCU {
		return Step{}, false
	}
	if state.Workspace.Limit < 0 || state.Workspace.Limit > CU(math.MaxInt64)-minimumNamespaceCU {
		return Step{}, false
	}
	to := from
	setAllocationActiveFixed(&to, state.Workspace.HA, currentFixed+minimumNamespaceCU)
	to.Limit += minimumNamespaceCU
	return Step{Action: ModifyWorkspaceFixed, From: from, To: to, Temporary: true}, true
}

func deleteNamespaceStep(state Tree, name string) Step {
	namespace := namespaceByName(state, name)
	return Step{Action: DeleteNamespace, Ref: Ref{Namespace: name}, From: capacityAllocation(*namespace.Capacity)}
}

func appendScheduledStep(state *Tree, steps *[]Step, step Step, phase string, budget *plannerBudget) error {
	if err := budget.addStep(phase); err != nil {
		return err
	}
	applyCandidate(state, candidate{step: step})
	*steps = append(*steps, step)
	return nil
}

func authoritativeStepIsSafe(state Tree, step Step, ceiling CU, phase string, budget *plannerBudget) (bool, error) {
	if isWorkspaceAction(step.Action) {
		if !workspaceStepFitsFOASInt32(step) || step.To.Limit > ceiling {
			return false, nil
		}
	}
	if err := budget.addWork(phase, capacityNodeCount(state)); err != nil {
		return false, err
	}
	next := cloneTree(state)
	applyCandidate(&next, candidate{step: step})
	return validatePlanningStateContext(budget.ctx, next) == nil, budget.checkContext()
}

func workspaceStepFitsFOASInt32(step Step) bool {
	fits := func(value CU) bool {
		return value >= 0 && value <= maximumFlinkSerializableComponentCU
	}
	switch step.Action {
	case ModifyWorkspaceFixed:
		return fits(step.To.FixedCU) && fits(step.To.CrossZoneFixedCU)
	case ModifyWorkspacePostpaid:
		return fits(step.To.Limit)
	case EnableWorkspaceElastic, ModifyWorkspaceElastic:
		elastic, ok := allocationElasticCU(step.To)
		return ok && fits(elastic)
	default:
		return true
	}
}

func allocationElasticCU(allocation Allocation) (CU, bool) {
	if allocation.FixedCU < 0 || allocation.CrossZoneFixedCU < 0 || allocation.Limit < 0 {
		return 0, false
	}
	if allocation.FixedCU > CU(math.MaxInt64)-allocation.CrossZoneFixedCU {
		return 0, false
	}
	totalFixed := allocation.FixedCU + allocation.CrossZoneFixedCU
	if allocation.Limit < totalFixed {
		return 0, false
	}
	return allocation.Limit - totalFixed, true
}

func isWorkspaceAction(action Action) bool {
	switch action {
	case ModifyWorkspaceFixed, ModifyWorkspacePostpaid, EnableWorkspaceElastic, ModifyWorkspaceElastic:
		return true
	default:
		return false
	}
}

func activeWorkspaceFixed(workspace WorkspaceCapacity) CU {
	if workspace.HA {
		return workspace.CrossZoneFixedCU
	}
	return workspace.FixedCU
}

func setAllocationActiveFixed(allocation *Allocation, ha bool, fixed CU) {
	if ha {
		allocation.FixedCU = 0
		allocation.CrossZoneFixedCU = fixed
		return
	}
	allocation.FixedCU = fixed
	allocation.CrossZoneFixedCU = 0
}

func movesToward(current, next, target CU) bool {
	if current < target {
		return next > current && next <= target
	}
	return next < current && next >= target
}

func usedLimitCU(used float64) CU {
	if used <= 0 {
		return 0
	}
	return CU(math.Ceil(used)) * 2
}

func clampNonNegative(value CU) CU {
	if value < 0 {
		return 0
	}
	return value
}

func minCU(first CU, rest ...CU) CU {
	result := first
	for _, value := range rest {
		if value < result {
			result = value
		}
	}
	return result
}

func maxCU(first CU, rest ...CU) CU {
	result := first
	for _, value := range rest {
		if value > result {
			result = value
		}
	}
	return result
}

func noSafeTransitionError() error {
	return fmt.Errorf("no safe capacity transition exists without exceeding endpoint capacity or violating child allocations")
}

func sortedNamespaces(input []Namespace) []Namespace {
	result := append([]Namespace(nil), input...)
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func newCandidate(action Action, ref Ref, from, to Allocation, level int) candidate {
	return candidate{
		step:      Step{Action: action, Ref: ref, From: from, To: to},
		level:     level,
		expansion: to.TotalFixed() > from.TotalFixed() || to.Limit > from.Limit || action == CreateNamespace,
	}
}

func betterLevel(candidate, selected int, expansion bool) bool {
	if expansion {
		return candidate < selected
	}
	return candidate > selected
}

func validatePlanningState(tree Tree) error {
	return validatePlanningStateContext(context.Background(), tree)
}

func validatePlanningStateContext(ctx context.Context, tree Tree) error {
	if err := validateObservedContext(ctx, tree); err != nil {
		return err
	}
	var fixed, crossZoneFixed, limit CU
	for _, namespace := range tree.Namespaces {
		if err := ctx.Err(); err != nil {
			return err
		}
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
