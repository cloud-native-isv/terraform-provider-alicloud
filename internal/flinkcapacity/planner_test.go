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
			{Name: "receiver", Capacity: capacity(firstFixed, firstLimit), Queues: []Queue{{Name: "q", Capacity: capacity(firstFixed, firstLimit)}}},
			{Name: "donor", Capacity: capacity(secondFixed, secondLimit), Queues: []Queue{{Name: "q", Capacity: capacity(secondFixed, secondLimit)}}},
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
		desired := twoNamespaceTree(24, 24, 12, 12, 4, 4)
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
	desired := plannerTree(8, 16, 8, 16, 4, 8)

	_, err := Plan(actual, desired)
	if err == nil || !strings.Contains(err.Error(), "used") {
		t.Fatalf("Plan() error = %v, want used capacity diagnostic", err)
	}
}

func TestPlanWorkspaceElasticActions(t *testing.T) {
	t.Run("zero to positive enables elastic", func(t *testing.T) {
		actual := plannerTree(8, 8, 8, 8, 8, 8)
		desired := plannerTree(8, 16, 8, 8, 8, 8)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) != 1 || steps[0].Action != EnableWorkspaceElastic {
			t.Fatalf("steps = %#v", steps)
		}
	})

	t.Run("positive to positive modifies elastic", func(t *testing.T) {
		actual := plannerTree(8, 16, 8, 8, 8, 8)
		desired := plannerTree(8, 20, 8, 8, 8, 8)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) != 1 || steps[0].Action != ModifyWorkspaceElastic {
			t.Fatalf("steps = %#v", steps)
		}
	})

	t.Run("positive to zero is rejected", func(t *testing.T) {
		actual := plannerTree(8, 16, 8, 8, 8, 8)
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

	steps, err := Plan(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].Action != ModifyWorkspacePostpaid {
		t.Fatalf("steps = %#v", steps)
	}
}

func TestPlanWorkspaceCompositionChange(t *testing.T) {
	t.Run("shrink elastic before expanding fixed when safe", func(t *testing.T) {
		actual := plannerTree(8, 16, 6, 12, 6, 12)
		desired := plannerTree(12, 16, 6, 12, 6, 12)
		steps, err := Plan(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) != 2 || steps[0].Action != ModifyWorkspaceElastic || steps[1].Action != ModifyWorkspaceFixed {
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
		desired := plannerTree(12, 16, 8, 16, 8, 16)
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
