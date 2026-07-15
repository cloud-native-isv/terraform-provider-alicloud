package flinkcapacity

import "fmt"

type Action string

const (
	ModifyWorkspaceFixed   Action = "modify_workspace_fixed"
	EnableWorkspaceElastic Action = "enable_workspace_elastic"
	ModifyWorkspaceElastic Action = "modify_workspace_elastic"
	ModifyNamespace        Action = "modify_namespace"
	ModifyQueue            Action = "modify_queue"
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
	Action Action
	Ref    Ref
	From   Allocation
	To     Allocation
}

type candidate struct {
	step      Step
	level     int
	expansion bool
}

// Plan returns a deterministic sequence of safe capacity changes. It never
// exceeds the larger endpoint workspace limit and only emits a step when the
// whole intermediate capacity tree remains valid.
func Plan(actual, desired Tree) ([]Step, error) {
	if actual.ChargeType != desired.ChargeType {
		return nil, fmt.Errorf("workspace charge type cannot change from %q to %q", actual.ChargeType, desired.ChargeType)
	}
	if err := validateSameTopology(actual, desired); err != nil {
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
	if err := validateUsedCapacity(state, target); err != nil {
		return nil, err
	}

	actualElastic := state.Workspace.AsCapacity().Elastic()
	desiredElastic := target.Workspace.AsCapacity().Elastic()
	if actualElastic > 0 && desiredElastic == 0 {
		return nil, fmt.Errorf("cannot reduce workspace elastic CU to zero through the supported public API; disable it manually before retrying")
	}

	maxWorkspaceLimit := state.Workspace.Limit
	if target.Workspace.Limit > maxWorkspaceLimit {
		maxWorkspaceLimit = target.Workspace.Limit
	}

	var steps []Step
	for attempts := 0; !sameCapacityTree(state, target); attempts++ {
		if attempts > capacityNodeCount(target)*4+8 {
			return nil, fmt.Errorf("capacity planner did not converge")
		}

		candidates := buildCandidates(state, target)
		selected := selectCandidate(state, candidates, maxWorkspaceLimit, true)
		if selected == nil {
			selected = selectCandidate(state, candidates, maxWorkspaceLimit, false)
		}
		if selected == nil {
			return nil, fmt.Errorf("no safe capacity transition exists without exceeding endpoint capacity or violating child allocations")
		}

		applyCandidate(&state, *selected)
		steps = append(steps, selected.step)
	}
	return steps, nil
}

func validateSameTopology(actual, desired Tree) error {
	desiredNamespaces := make(map[string]Namespace, len(desired.Namespaces))
	for _, namespace := range desired.Namespaces {
		desiredNamespaces[namespace.Name] = namespace
	}
	actualNamespaces := make(map[string]Namespace, len(actual.Namespaces))
	for _, namespace := range actual.Namespaces {
		actualNamespaces[namespace.Name] = namespace
		desiredNamespace, ok := desiredNamespaces[namespace.Name]
		if !ok {
			return fmt.Errorf("cloud namespace %q is undeclared", namespace.Name)
		}
		desiredQueues := make(map[string]struct{}, len(desiredNamespace.Queues))
		for _, queue := range desiredNamespace.Queues {
			desiredQueues[queue.Name] = struct{}{}
		}
		for _, queue := range namespace.Queues {
			if _, ok := desiredQueues[queue.Name]; !ok {
				return fmt.Errorf("cloud queue %q/%q is undeclared", namespace.Name, queue.Name)
			}
		}
	}
	for _, namespace := range desired.Namespaces {
		actualNamespace, ok := actualNamespaces[namespace.Name]
		if !ok {
			return fmt.Errorf("declared namespace %q does not exist in the workspace", namespace.Name)
		}
		actualQueues := make(map[string]struct{}, len(actualNamespace.Queues))
		for _, queue := range actualNamespace.Queues {
			actualQueues[queue.Name] = struct{}{}
		}
		for _, queue := range namespace.Queues {
			if _, ok := actualQueues[queue.Name]; !ok {
				return fmt.Errorf("declared queue %q/%q does not exist in the workspace", namespace.Name, queue.Name)
			}
		}
	}
	return nil
}

func validateUsedCapacity(actual, desired Tree) error {
	for _, desiredNamespace := range desired.Namespaces {
		actualNamespace := namespaceByName(actual, desiredNamespace.Name)
		if actualNamespace.Used > desiredNamespace.Capacity.Limit {
			return fmt.Errorf("namespace %q desired limit %v is below used CU %v", desiredNamespace.Name, desiredNamespace.Capacity.Limit.Float64(), actualNamespace.Used.Float64())
		}
		for _, desiredQueue := range desiredNamespace.Queues {
			actualQueue := queueByName(*actualNamespace, desiredQueue.Name)
			if actualQueue.Used > desiredQueue.Capacity.Limit {
				return fmt.Errorf("queue %q/%q desired limit %v is below used CU %v", desiredNamespace.Name, desiredQueue.Name, desiredQueue.Capacity.Limit.Float64(), actualQueue.Used.Float64())
			}
		}
	}
	return nil
}

func buildCandidates(state, target Tree) []candidate {
	result := make([]candidate, 0, capacityNodeCount(target)+1)
	currentWorkspace := workspaceAllocation(state.Workspace)
	targetWorkspace := workspaceAllocation(target.Workspace)
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

	for _, targetNamespace := range target.Namespaces {
		currentNamespace := namespaceByName(state, targetNamespace.Name)
		from := capacityAllocation(*currentNamespace.Capacity)
		to := capacityAllocation(*targetNamespace.Capacity)
		if from != to {
			result = append(result, newCandidate(ModifyNamespace, Ref{Namespace: targetNamespace.Name}, from, to, 1))
		}
		for _, targetQueue := range targetNamespace.Queues {
			currentQueue := queueByName(*currentNamespace, targetQueue.Name)
			from = capacityAllocation(*currentQueue.Capacity)
			to = capacityAllocation(*targetQueue.Capacity)
			if from != to {
				result = append(result, newCandidate(ModifyQueue, Ref{Namespace: targetNamespace.Name, Queue: targetQueue.Name}, from, to, 2))
			}
		}
	}
	return result
}

func newCandidate(action Action, ref Ref, from, to Allocation, level int) candidate {
	return candidate{
		step:      Step{Action: action, Ref: ref, From: from, To: to},
		level:     level,
		expansion: to.TotalFixed() > from.TotalFixed() || to.Limit > from.Limit,
	}
}

func selectCandidate(state Tree, candidates []candidate, maxWorkspaceLimit CU, expansion bool) *candidate {
	selectedIndex := -1
	for i := range candidates {
		candidate := candidates[i]
		if candidate.expansion != expansion || !candidateIsSafe(state, candidate, maxWorkspaceLimit) {
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

func candidateIsSafe(state Tree, candidate candidate, maxWorkspaceLimit CU) bool {
	if candidate.step.To.Limit > maxWorkspaceLimit && candidate.level == 0 {
		return false
	}
	next := cloneTree(state)
	applyCandidate(&next, candidate)
	_, err := Resolve(next)
	return err == nil
}

func applyCandidate(state *Tree, candidate candidate) {
	to := candidate.step.To
	switch candidate.step.Action {
	case ModifyWorkspaceFixed, EnableWorkspaceElastic, ModifyWorkspaceElastic:
		state.Workspace.FixedCU = to.FixedCU
		state.Workspace.CrossZoneFixedCU = to.CrossZoneFixedCU
		state.Workspace.Limit = to.Limit
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
	if actual.Workspace != desired.Workspace {
		return false
	}
	for _, desiredNamespace := range desired.Namespaces {
		actualNamespace := namespaceByName(actual, desiredNamespace.Name)
		if *actualNamespace.Capacity != *desiredNamespace.Capacity {
			return false
		}
		for _, desiredQueue := range desiredNamespace.Queues {
			actualQueue := queueByName(*actualNamespace, desiredQueue.Name)
			if *actualQueue.Capacity != *desiredQueue.Capacity {
				return false
			}
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
