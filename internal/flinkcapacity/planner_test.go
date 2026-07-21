package flinkcapacity

import (
	"strings"
	"testing"
)

func plannerTree(workspaceFixed, workspaceLimit, namespaceFixed, namespaceLimit, queueFixed, queueLimit CU) Tree {
	return Tree{
		ChargeType: "PRE",
		Workspace: WorkspaceCapacity{
			FixedCU: workspaceFixed,
			Limit:   workspaceLimit,
		},
		Namespaces: []Namespace{
			{
				Name:     "default",
				Capacity: capacity(namespaceFixed, namespaceLimit),
				Queues: []Queue{
					{Name: "default-queue", Capacity: capacity(queueFixed, queueLimit)},
				},
			},
		},
	}
}

func stepLevel(step Step) string {
	switch step.Action {
	case ModifyQueue:
		return "queue"
	case ModifyNamespace:
		return "namespace"
	default:
		return "workspace"
	}
}

func firstIndexOfLevel(steps []Step, level string) int {
	for i, step := range steps {
		if stepLevel(step) == level {
			return i
		}
	}
	return -1
}

func lastIndexOfLevel(steps []Step, level string) int {
	for i := len(steps) - 1; i >= 0; i-- {
		if stepLevel(steps[i]) == level {
			return i
		}
	}
	return -1
}

func TestPlanHierarchyOrdering(t *testing.T) {
	t.Run("expand parent before child", func(t *testing.T) {
		actual := plannerTree(8, 16, 8, 16, 8, 16)
		desired := plannerTree(16, 32, 16, 32, 16, 32)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if lastIndexOfLevel(steps, "workspace") >= firstIndexOfLevel(steps, "namespace") {
			t.Fatalf("workspace expansion must finish before namespace: %#v", steps)
		}
		if lastIndexOfLevel(steps, "namespace") >= firstIndexOfLevel(steps, "queue") {
			t.Fatalf("namespace expansion must finish before queue: %#v", steps)
		}
	})

	t.Run("shrink child before parent", func(t *testing.T) {
		actual := plannerTree(16, 32, 16, 32, 16, 32)
		desired := plannerTree(8, 16, 8, 16, 8, 16)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if lastIndexOfLevel(steps, "queue") >= firstIndexOfLevel(steps, "namespace") {
			t.Fatalf("queue shrink must finish before namespace: %#v", steps)
		}
		if lastIndexOfLevel(steps, "namespace") >= firstIndexOfLevel(steps, "workspace") {
			t.Fatalf("namespace shrink must finish before workspace: %#v", steps)
		}
	})
}

func TestPlanRejectsUndeclaredCloudChildren(t *testing.T) {
	actual := plannerTree(8, 16, 8, 16, 8, 16)
	actual.Namespaces[0].Queues = append(actual.Namespaces[0].Queues, Queue{
		Name:     "undeclared",
		Capacity: capacity(1, 1),
	})
	desired := plannerTree(8, 16, 8, 16, 8, 16)

	_, err := Plan(actual, desired)
	if err == nil || !strings.Contains(err.Error(), "undeclared") {
		t.Fatalf("Plan() error = %v, want undeclared queue name", err)
	}
}

func TestPlanRejectsHAConversion(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	desired := cloneTree(actual)
	actual.Workspace.HA = true
	actual.Workspace.FixedCU = 0
	actual.Workspace.CrossZoneFixedCU = 8

	_, err := Plan(actual, desired)
	if err == nil || !strings.Contains(err.Error(), "high availability") {
		t.Fatalf("Plan() error = %v", err)
	}
}

func twoNamespaceTree(parentFixed, parentLimit, firstFixed, firstLimit, secondFixed, secondLimit CU) Tree {
	return Tree{
		ChargeType: "PRE",
		Workspace:  WorkspaceCapacity{FixedCU: parentFixed, Limit: parentLimit},
		Namespaces: []Namespace{
			{Name: "receiver", Capacity: capacity(firstFixed, firstLimit), Queues: []Queue{{Name: "default-queue", Capacity: capacity(firstFixed, firstLimit)}}},
			{Name: "donor", Capacity: capacity(secondFixed, secondLimit), Queues: []Queue{{Name: "default-queue", Capacity: capacity(secondFixed, secondLimit)}}},
		},
	}
}

func indexOfNamespaceStep(steps []Step, namespace string) int {
	for i, step := range steps {
		if step.Ref.Namespace == namespace && step.Action == ModifyNamespace {
			return i
		}
	}
	return -1
}

func TestPlanRebalance(t *testing.T) {
	t.Run("receiver first when parent has headroom", func(t *testing.T) {
		actual := twoNamespaceTree(24, 24, 8, 8, 8, 8)
		desired := twoNamespaceTree(16, 16, 12, 12, 4, 4)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if indexOfNamespaceStep(steps, "receiver") >= indexOfNamespaceStep(steps, "donor") {
			t.Fatalf("receiver should expand before donor shrinks: %#v", steps)
		}
	})

	t.Run("donor first when parent has no headroom", func(t *testing.T) {
		actual := twoNamespaceTree(16, 16, 8, 8, 8, 8)
		desired := twoNamespaceTree(16, 16, 12, 12, 4, 4)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if indexOfNamespaceStep(steps, "donor") >= indexOfNamespaceStep(steps, "receiver") {
			t.Fatalf("donor should shrink before receiver expands: %#v", steps)
		}
	})
}

func TestPlanUsedGate(t *testing.T) {
	actual := plannerTree(8, 16, 8, 16, 8, 16)
	actual.Namespaces[0].Queues[0].Used = 12
	desired := plannerTree(8, 8, 8, 8, 8, 8)

	_, err := Plan(actual, desired)
	if err == nil || !strings.Contains(err.Error(), "used") {
		t.Fatalf("Plan() error = %v, want used capacity diagnostic", err)
	}
}

func TestPlanUsedGateRejectsFractionalUsageAtEveryLevel(t *testing.T) {
	tests := []struct {
		name string
		set  func(*Tree)
	}{
		{
			name: "workspace",
			set: func(tree *Tree) {
				tree.Workspace.Used = 4.1
			},
		},
		{
			name: "namespace",
			set: func(tree *Tree) {
				tree.Namespaces[0].Used = 4.1
			},
		},
		{
			name: "queue",
			set: func(tree *Tree) {
				tree.Namespaces[0].Queues[0].Used = 4.1
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := plannerTree(8, 16, 8, 16, 8, 16)
			desired := plannerTree(8, 8, 8, 8, 8, 8)
			tc.set(&actual)

			_, err := Plan(actual, desired)
			if err == nil || !strings.Contains(err.Error(), "below used CU 4.1") {
				t.Fatalf("Plan() error = %v, want exact fractional used capacity diagnostic", err)
			}
		})
	}
}

func TestPlanWorkspaceElasticActions(t *testing.T) {
	t.Run("zero to positive enables elastic", func(t *testing.T) {
		actual := plannerTree(8, 8, 8, 8, 8, 8)
		desired := plannerTree(8, 16, 8, 16, 8, 16)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) == 0 || steps[0].Action != EnableWorkspaceElastic {
			t.Fatalf("steps = %#v", steps)
		}
	})

	t.Run("positive to positive modifies elastic", func(t *testing.T) {
		actual := plannerTree(8, 16, 8, 16, 8, 16)
		desired := plannerTree(8, 20, 8, 20, 8, 20)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) == 0 || steps[0].Action != ModifyWorkspaceElastic {
			t.Fatalf("steps = %#v", steps)
		}
	})

	t.Run("positive to zero is rejected", func(t *testing.T) {
		actual := plannerTree(8, 16, 8, 16, 8, 16)
		desired := plannerTree(8, 8, 8, 8, 8, 8)
		_, err := Plan(actual, desired)
		if err == nil || !strings.Contains(err.Error(), "cannot reduce workspace elastic CU to zero") {
			t.Fatalf("Plan() error = %v", err)
		}
	})
}

func TestPlanPostpaidWorkspaceUsesInstanceSpecAction(t *testing.T) {
	actual := Tree{
		ChargeType: "POST",
		Workspace:  WorkspaceCapacity{Limit: 8},
		Namespaces: []Namespace{{
			Name:     "default",
			Capacity: capacity(0, 8),
			Queues:   []Queue{{Name: "default-queue", Capacity: capacity(0, 8)}},
		}},
	}
	desired := cloneTree(actual)
	desired.Workspace.Limit = 16
	desired.Namespaces[0].Capacity = capacity(0, 16)
	desired.Namespaces[0].Queues[0].Capacity = capacity(0, 16)

	steps, err := Plan(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) == 0 || steps[0].Action != ModifyWorkspacePostpaid {
		t.Fatalf("steps = %#v", steps)
	}
}

func TestPlanPostpaidChildStepsRemainPureElastic(t *testing.T) {
	actual := Tree{
		ChargeType: "POST",
		Workspace:  WorkspaceCapacity{Limit: 8},
		Namespaces: []Namespace{{
			Name:     "default",
			Capacity: capacity(0, 4),
			Queues:   []Queue{{Name: "default-queue", Capacity: capacity(0, 4)}},
		}},
	}
	desired := cloneTree(actual)
	desired.Workspace.Limit = 8
	desired.Namespaces[0].Capacity = capacity(0, 8)
	desired.Namespaces[0].Queues[0].Capacity = capacity(0, 8)

	steps, err := Plan(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 {
		t.Fatalf("steps = %#v", steps)
	}
	for _, step := range steps {
		if step.To.FixedCU != 0 || step.To.Limit != 8 {
			t.Fatalf("step = %#v, want pure elastic child allocation", step)
		}
	}
}

func TestPlanWorkspaceCompositionChange(t *testing.T) {
	t.Run("shrink elastic before expanding fixed when safe", func(t *testing.T) {
		actual := plannerTree(8, 16, 6, 12, 6, 12)
		desired := plannerTree(12, 16, 12, 16, 12, 16)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) < 2 || steps[0].Action != ModifyWorkspaceElastic || steps[1].Action != ModifyWorkspaceFixed {
			t.Fatalf("steps = %#v", steps)
		}
		for _, step := range steps {
			if step.To.Limit > 16 {
				t.Fatalf("temporary capacity exceeds endpoint limit: %#v", steps)
			}
		}
	})

	t.Run("reject when neither order is safe", func(t *testing.T) {
		actual := plannerTree(8, 16, 8, 16, 8, 16)
		desired := plannerTree(12, 16, 12, 16, 12, 16)
		_, err := Plan(actual, desired)
		if err == nil || !strings.Contains(err.Error(), "no safe capacity transition") {
			t.Fatalf("Plan() error = %v", err)
		}
	})
}

func TestPlanIsDeterministic(t *testing.T) {
	actual := twoNamespaceTree(16, 16, 8, 8, 8, 8)
	desired := twoNamespaceTree(16, 16, 12, 12, 4, 4)
	first, err := Plan(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Plan(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(second) {
		t.Fatalf("nondeterministic lengths: %d != %d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("step %d differs: %#v != %#v", i, first[i], second[i])
		}
	}
}

func TestPlanTopologyDeletionOrderIsIndependentOfObservedOrder(t *testing.T) {
	first := authoritativeTree(6,
		authoritativeNamespace("keep", 2, 2, 0),
		authoritativeNamespace("drop-a", 2, 2, 0),
		authoritativeNamespace("drop-b", 2, 2, 0),
	)
	second := cloneTree(first)
	second.Namespaces[1], second.Namespaces[2] = second.Namespaces[2], second.Namespaces[1]
	desired := authoritativeTree(2, authoritativeNamespace("keep", 2, 2, 0))

	firstSteps, err := PlanAuthoritative(first, desired)
	if err != nil {
		t.Fatal(err)
	}
	secondSteps, err := PlanAuthoritative(second, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(firstSteps) != len(secondSteps) {
		t.Fatalf("step counts differ: %#v != %#v", firstSteps, secondSteps)
	}
	for i := range firstSteps {
		if firstSteps[i] != secondSteps[i] {
			t.Fatalf("step %d differs: %#v != %#v", i, firstSteps[i], secondSteps[i])
		}
	}
}

func authoritativeTree(workspace CU, namespaces ...Namespace) Tree {
	return Tree{
		ChargeType: "PRE",
		Workspace:  WorkspaceCapacity{FixedCU: workspace, Limit: workspace},
		Namespaces: namespaces,
	}
}

func authoritativeNamespace(name string, fixed, limit CU, used float64) Namespace {
	return Namespace{
		Name:     name,
		Capacity: capacity(fixed, limit),
		Used:     used,
		Queues: []Queue{{
			Name:     "default-queue",
			Capacity: capacity(fixed, limit),
		}},
	}
}

func namespaceWithQueues(name string, fixed, limit CU, queues ...Queue) Namespace {
	return Namespace{
		Name:     name,
		Capacity: capacity(fixed, limit),
		Queues:   queues,
	}
}

func TestPlanRenamesOnlyNamespaceWithOneCUTemporaryBuffer(t *testing.T) {
	actual := authoritativeTree(4, authoritativeNamespace("old", 4, 4, 0))
	desired := authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0))

	steps, err := PlanAuthoritative(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	wantActions := []Action{
		ModifyWorkspaceFixed,
		CreateNamespace,
		DeleteNamespace,
	}
	if len(steps) != len(wantActions) {
		t.Fatalf("steps = %#v", steps)
	}
	peak := CU(0)
	for i, step := range steps {
		if step.Action != wantActions[i] {
			t.Fatalf("step %d action = %q, want %q; all steps = %#v", i, step.Action, wantActions[i], steps)
		}
		if step.Action == CreateNamespace && (step.To.FixedCU != minimumNamespaceCU || step.To.Limit != minimumNamespaceCU) {
			t.Fatalf("create allocation = %#v, want exactly one CU", step.To)
		}
		if step.Action == ModifyWorkspaceFixed && step.To.TotalFixed() > peak {
			peak = step.To.TotalFixed()
		}
	}
	if peak != 6 { // 3 CU: endpoint 2 CU plus exactly 1 CU temporary capacity.
		t.Fatalf("peak workspace fixed = %v, want 6 half-CU", peak)
	}
}

func TestPlanRejectsCustomQueueInRetainedNamespaceBeforeWrites(t *testing.T) {
	actual := authoritativeTree(4, namespaceWithQueues(
		"keep", 4, 4,
		Queue{Name: "default-queue", Capacity: capacity(4, 4)},
		Queue{Name: "custom", Capacity: capacity(1, 2), Used: 0.25},
	))
	desired := authoritativeTree(4, authoritativeNamespace("keep", 4, 4, 0))

	_, err := PlanAuthoritative(actual, desired)
	want := `retained namespace "keep" contains unsupported queue "custom" (fixed_cu=0.5, max_cu_limit=1, used_cu=0.25)`
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("Plan() error = %v", err)
	}
}

func TestPlanDeletionIsAReadBarrierBeforeWorkspaceUsedGate(t *testing.T) {
	actual := authoritativeTree(6,
		authoritativeNamespace("keep", 2, 2, 0.5),
		authoritativeNamespace("drop", 4, 4, 2.5),
	)
	actual.Workspace.Used = 3
	desired := authoritativeTree(2, authoritativeNamespace("keep", 2, 2, 0))

	steps, err := PlanAuthoritative(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].Action != DeleteNamespace || steps[0].Ref.Namespace != "drop" {
		t.Fatalf("steps = %#v, want only the delete read barrier", steps)
	}

	refreshed := cloneTree(actual)
	applyCandidate(&refreshed, candidate{step: steps[0]})
	refreshed.Workspace.Used = 0.5
	steps, err = PlanAuthoritative(refreshed, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) == 0 || steps[0].Action != ModifyWorkspaceFixed {
		t.Fatalf("post-delete steps = %#v, want workspace shrink after refreshed used", steps)
	}
}

func TestPlanResumesMultiRenameAfterCreateWithoutStackingBuffer(t *testing.T) {
	actual := authoritativeTree(6,
		authoritativeNamespace("old-a", 2, 2, 0),
		authoritativeNamespace("old-b", 2, 2, 0),
		authoritativeNamespace("new-a", 2, 2, 0),
	)
	desired := authoritativeTree(4,
		authoritativeNamespace("new-a", 2, 2, 0),
		authoritativeNamespace("new-b", 2, 2, 0),
	)

	steps, err := PlanAuthoritative(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].Action != DeleteNamespace || steps[0].Ref.Namespace != "old-a" {
		t.Fatalf("steps = %#v, want old-a deletion barrier using the existing buffer", steps)
	}
}

func TestPlanSupportsNamespaceFixedShrinkWithLimitExpansion(t *testing.T) {
	actual := authoritativeTree(6,
		authoritativeNamespace("a", 4, 8, 0),
		authoritativeNamespace("b", 2, 10, 0),
	)
	actual.Workspace.Limit = 18
	desired := authoritativeTree(6,
		authoritativeNamespace("a", 2, 10, 0),
		authoritativeNamespace("b", 4, 8, 0),
	)
	desired.Workspace.Limit = 18

	steps, err := PlanAuthoritative(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	got := cloneTree(actual)
	for _, step := range steps {
		if step.Action == ModifyNamespace || step.Action == ModifyQueue {
			fixedChanged := step.From.TotalFixed() != step.To.TotalFixed()
			limitChanged := step.From.Limit != step.To.Limit
			if fixedChanged && limitChanged {
				t.Fatalf("child step changed fixed and limit atomically: %#v", step)
			}
		}
		applyCandidate(&got, candidate{step: step})
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("steps did not reach crossed fixed/limit allocation: %#v", steps)
	}
}

func TestPlanReservesExistingHeadroomForCreateBeforeRetainedExpansion(t *testing.T) {
	actual := authoritativeTree(8,
		authoritativeNamespace("a-retained", 2, 2, 0),
		authoritativeNamespace("m-surplus", 4, 4, 0),
	)
	desired := authoritativeTree(8,
		authoritativeNamespace("a-retained", 4, 4, 0),
		authoritativeNamespace("z-new", 4, 4, 0),
	)

	for _, reverse := range []bool{false, true} {
		observed := cloneTree(actual)
		wanted := cloneTree(desired)
		if reverse {
			observed.Namespaces[0], observed.Namespaces[1] = observed.Namespaces[1], observed.Namespaces[0]
			wanted.Namespaces[0], wanted.Namespaces[1] = wanted.Namespaces[1], wanted.Namespaces[0]
		}
		steps, err := PlanAuthoritative(observed, wanted)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) == 0 || steps[0].Action != CreateNamespace || steps[0].Ref.Namespace != "z-new" {
			t.Fatalf("reverse=%v steps=%#v, want topology create before retained expansion", reverse, steps)
		}
		if steps[0].Temporary {
			t.Fatalf("reverse=%v used temporary capacity despite existing headroom: %#v", reverse, steps)
		}
	}
}

func TestPlanDeletesSurplusBeforeCreateWhenThatAvoidsTemporaryCapacity(t *testing.T) {
	actual := authoritativeTree(4,
		authoritativeNamespace("old", 2, 2, 0),
		authoritativeNamespace("surplus", 2, 2, 0),
	)
	desired := authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0))

	steps, err := PlanAuthoritative(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].Action != DeleteNamespace || steps[0].Temporary {
		t.Fatalf("steps = %#v, want surplus deletion barrier without temporary capacity", steps)
	}
}

func TestPlanAllowsDeletingBusyUndeclaredNamespace(t *testing.T) {
	actual := authoritativeTree(4,
		authoritativeNamespace("keep", 2, 2, 0),
		namespaceWithQueues(
			"drop", 2, 2,
			Queue{Name: "default-queue", Capacity: capacity(2, 2), Used: 1.5},
			Queue{Name: "custom", Capacity: capacity(1, 1), Used: 0.5},
		),
	)
	desired := authoritativeTree(4, authoritativeNamespace("keep", 4, 4, 0))

	steps, err := PlanAuthoritative(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range steps {
		if step.Action == DeleteNamespace && step.Ref.Namespace == "drop" {
			return
		}
	}
	t.Fatalf("steps = %#v, want deletion of undeclared namespace", steps)
}

func TestValidateDesiredRequiresExactFixedAndLimitSums(t *testing.T) {
	t.Run("fixed", func(t *testing.T) {
		desired := authoritativeTree(4, authoritativeNamespace("only", 2, 4, 0))
		err := ValidateDesired(desired)
		if err == nil || !strings.Contains(err.Error(), "namespace fixed CU sum 1 must equal workspace fixed CU 2") {
			t.Fatalf("ValidateDesired() error = %v", err)
		}
	})
	t.Run("max", func(t *testing.T) {
		desired := authoritativeTree(4, authoritativeNamespace("only", 4, 4, 0))
		desired.Workspace.Limit = 6
		err := ValidateDesired(desired)
		if err == nil || !strings.Contains(err.Error(), "namespace max CU sum 2 must equal workspace max CU 3") {
			t.Fatalf("ValidateDesired() error = %v", err)
		}
	})
}

func TestValidateDesiredRequiresAtLeastOneCUPerNamespace(t *testing.T) {
	desired := authoritativeTree(minimumNamespaceCU-1,
		authoritativeNamespace("too-small", minimumNamespaceCU-1, minimumNamespaceCU-1, 0),
	)

	err := ValidateDesired(desired)
	if err == nil || !strings.Contains(err.Error(), "at least 1") {
		t.Fatalf("ValidateDesired() error = %v, want one-CU namespace minimum", err)
	}
}

func TestPlanAuthoritativeRejectsPostpaidWorkspace(t *testing.T) {
	actual := postpaidAuthoritativeTree(8)

	_, err := PlanAuthoritative(actual, cloneTree(actual))
	if err == nil || !strings.Contains(err.Error(), "PRE") {
		t.Fatalf("PlanAuthoritative() error = %v, want PRE-only error", err)
	}
}

func postpaidAuthoritativeTree(limit CU) Tree {
	return Tree{
		ChargeType: "POST",
		Workspace:  WorkspaceCapacity{Limit: limit},
		Namespaces: []Namespace{authoritativeNamespace("default", 0, limit, 0)},
	}
}
