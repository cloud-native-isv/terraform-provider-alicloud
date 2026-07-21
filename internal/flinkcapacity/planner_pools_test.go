package flinkcapacity

import (
	"strings"
	"testing"
)

func setNamespaceCrossZone(t *testing.T, namespace *Namespace, crossZone bool) {
	t.Helper()
	namespace.CrossZone = crossZone
}

func namespaceCrossZone(t *testing.T, namespace Namespace) bool {
	t.Helper()
	return namespace.CrossZone
}

func stepNamespaceCrossZone(t *testing.T, step Step) bool {
	t.Helper()
	return step.NamespaceCrossZone
}

func setStepNamespaceCrossZone(t *testing.T, step *Step, crossZone bool) {
	t.Helper()
	step.NamespaceCrossZone = crossZone
}

func TestAuthoritativePlannerPoolAwareValidation(t *testing.T) {
	t.Run("accepts pure single-zone and HA pools", func(t *testing.T) {
		single := authoritativeTree(4, authoritativeNamespace("single", 4, 4, 0))
		setNamespaceCrossZone(t, &single.Namespaces[0], false)
		if err := ValidateDesired(single); err != nil {
			t.Fatalf("single-zone ValidateDesired() error = %v", err)
		}

		ha := Tree{
			ChargeType: "PRE",
			Workspace:  WorkspaceCapacity{HA: true, CrossZoneFixedCU: 4, Limit: 4},
			Namespaces: []Namespace{authoritativeNamespace("ha", 4, 4, 0)},
		}
		setNamespaceCrossZone(t, &ha.Namespaces[0], true)
		if err := ValidateDesired(ha); err != nil {
			t.Fatalf("HA ValidateDesired() error = %v", err)
		}
	})

	t.Run("rejects mixed workspace and namespace pools", func(t *testing.T) {
		mixedWorkspace := authoritativeTree(4, authoritativeNamespace("only", 4, 4, 0))
		mixedWorkspace.Workspace.HA = true
		mixedWorkspace.Workspace.FixedCU = 2
		mixedWorkspace.Workspace.CrossZoneFixedCU = 2
		if err := ValidateDesired(mixedWorkspace); err == nil || !strings.Contains(err.Error(), "pure") {
			t.Fatalf("mixed workspace ValidateDesired() error = %v, want pure-mode rejection", err)
		}

		mixedNamespace := authoritativeTree(4, authoritativeNamespace("only", 4, 4, 0))
		setNamespaceCrossZone(t, &mixedNamespace.Namespaces[0], true)
		if err := ValidateDesired(mixedNamespace); err == nil || !strings.Contains(err.Error(), "cross-zone") {
			t.Fatalf("mixed namespace ValidateDesired() error = %v, want cross-zone rejection", err)
		}

		nonHAWithCrossZonePool := authoritativeTree(4, authoritativeNamespace("only", 4, 4, 0))
		nonHAWithCrossZonePool.Workspace.FixedCU = 2
		nonHAWithCrossZonePool.Workspace.CrossZoneFixedCU = 2
		if err := ValidateDesired(nonHAWithCrossZonePool); err == nil {
			t.Fatal("non-HA workspace with cross-zone fixed pool was accepted")
		}

		withoutFixedPool := authoritativeTree(0, authoritativeNamespace("only", 2, 2, 0))
		withoutFixedPool.Workspace.Limit = 2
		if err := ValidateDesired(withoutFixedPool); err == nil {
			t.Fatal("workspace without a fixed pool was accepted")
		}
	})

	t.Run("conserves cross-zone fixed CU separately", func(t *testing.T) {
		ha := Tree{
			ChargeType: "PRE",
			Workspace:  WorkspaceCapacity{HA: true, CrossZoneFixedCU: 4, Limit: 4},
			Namespaces: []Namespace{authoritativeNamespace("ha", 2, 2, 0)},
		}
		setNamespaceCrossZone(t, &ha.Namespaces[0], true)
		if err := ValidateDesired(ha); err == nil || !strings.Contains(err.Error(), "cross-zone fixed CU sum") {
			t.Fatalf("HA pool conservation error = %v, want cross-zone fixed sum", err)
		}
	})

	t.Run("rejects fractional authoritative allocations", func(t *testing.T) {
		fractional := authoritativeTree(3, authoritativeNamespace("only", 3, 3, 0))
		if err := ValidateDesired(fractional); err == nil || !strings.Contains(err.Error(), "integer") {
			t.Fatalf("fractional ValidateDesired() error = %v, want integer-CU rejection", err)
		}
	})

	t.Run("rejects retained namespace type mismatch before writes", func(t *testing.T) {
		actual := authoritativeTree(4, authoritativeNamespace("retained", 4, 4, 0))
		desired := cloneTree(actual)
		setNamespaceCrossZone(t, &actual.Namespaces[0], true)
		setNamespaceCrossZone(t, &desired.Namespaces[0], false)
		_, err := PlanAuthoritative(actual, desired)
		if err == nil || !strings.Contains(err.Error(), "type") {
			t.Fatalf("PlanAuthoritative() error = %v, want retained type mismatch", err)
		}
	})
}

func TestAuthoritativePlannerUsesCorrectPoolForTemporaryNamespaceCapacity(t *testing.T) {
	t.Run("single-zone rename expands only the primary pool", func(t *testing.T) {
		actual := authoritativeTree(4, authoritativeNamespace("old", 4, 4, 0))
		desired := authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0))
		steps, err := PlanAuthoritative(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) < 2 || steps[0].Action != ModifyWorkspaceFixed || steps[0].To.FixedCU != 6 || steps[0].To.CrossZoneFixedCU != 0 || !steps[0].Temporary {
			t.Fatalf("single-zone temporary step = %#v", steps)
		}
	})

	t.Run("HA rename expands only the cross-zone pool and types create", func(t *testing.T) {
		actual := Tree{
			ChargeType: "PRE",
			Workspace:  WorkspaceCapacity{HA: true, CrossZoneFixedCU: 4, Limit: 4},
			Namespaces: []Namespace{authoritativeNamespace("old", 4, 4, 0)},
		}
		desired := cloneTree(actual)
		desired.Namespaces[0].Name = "new"
		setNamespaceCrossZone(t, &actual.Namespaces[0], true)
		setNamespaceCrossZone(t, &desired.Namespaces[0], true)

		steps, err := PlanAuthoritative(actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if len(steps) < 2 || steps[0].Action != ModifyWorkspaceFixed || steps[0].To.FixedCU != 0 || steps[0].To.CrossZoneFixedCU != 6 || steps[0].To.Limit != 6 || !steps[0].Temporary {
			t.Fatalf("HA temporary step = %#v", steps)
		}
		for _, step := range steps {
			if step.Action == CreateNamespace {
				if !stepNamespaceCrossZone(t, step) {
					t.Fatalf("HA create step must be cross-zone: %#v", step)
				}
				return
			}
		}
		t.Fatalf("steps = %#v, want create namespace", steps)
	})
}

func TestAuthoritativePlannerDoesNotStackHARecoveryBuffer(t *testing.T) {
	actual := Tree{
		ChargeType: "PRE",
		Workspace:  WorkspaceCapacity{HA: true, CrossZoneFixedCU: 6, Limit: 6},
		Namespaces: []Namespace{
			authoritativeNamespace("old-a", 2, 2, 0),
			authoritativeNamespace("old-b", 2, 2, 0),
			authoritativeNamespace("new-a", 2, 2, 0),
		},
	}
	desired := Tree{
		ChargeType: "PRE",
		Workspace:  WorkspaceCapacity{HA: true, CrossZoneFixedCU: 4, Limit: 4},
		Namespaces: []Namespace{
			authoritativeNamespace("new-a", 2, 2, 0),
			authoritativeNamespace("new-b", 2, 2, 0),
		},
	}
	for i := range actual.Namespaces {
		setNamespaceCrossZone(t, &actual.Namespaces[i], true)
	}
	for i := range desired.Namespaces {
		setNamespaceCrossZone(t, &desired.Namespaces[i], true)
	}

	steps, err := PlanAuthoritative(actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].Action != DeleteNamespace || steps[0].Ref.Namespace != "old-a" || steps[0].Temporary {
		t.Fatalf("steps = %#v, want existing HA buffer to avoid another temporary expansion", steps)
	}
}

func TestAuthoritativeApplyAndComparisonRetainNamespaceType(t *testing.T) {
	tree := authoritativeTree(2, authoritativeNamespace("old", 2, 2, 0))
	step := Step{Action: CreateNamespace, Ref: Ref{Namespace: "new"}, To: Allocation{FixedCU: 2, Limit: 2}}
	setStepNamespaceCrossZone(t, &step, true)
	applyCandidate(&tree, candidate{step: step})
	created := namespaceByName(tree, "new")
	if created == nil || !namespaceCrossZone(t, *created) {
		t.Fatalf("created namespace = %#v, want cross-zone type", created)
	}
	if !stepConverged(tree, step) {
		t.Fatalf("stepConverged() rejected created cross-zone namespace: %#v", tree)
	}
	tree.Namespaces[1].CrossZone = false
	if stepConverged(tree, step) {
		t.Fatalf("stepConverged() ignored created namespace type: %#v", tree)
	}
	tree.Namespaces[1].CrossZone = true
	desired := cloneTree(tree)
	setNamespaceCrossZone(t, &desired.Namespaces[1], false)
	if sameCapacityTree(tree, desired) {
		t.Fatal("sameCapacityTree() ignored namespace type")
	}
}
