package flinkcapacity

import (
	"strings"
	"testing"
)

func capacity(fixed, limit CU) *Capacity {
	return &Capacity{Fixed: fixed, Limit: limit}
}

func baseTree() Tree {
	return Tree{
		ChargeType: "PRE",
		Workspace: WorkspaceCapacity{
			HA:               true,
			CrossZoneFixedCU: 32,
			Limit:            64,
		},
		Namespaces: []Namespace{
			{
				Name: "default",
				Queues: []Queue{
					{Name: "default-queue"},
				},
			},
		},
	}
}

func TestResolveSingleRemainderGetsAll(t *testing.T) {
	resolved, err := Resolve(baseTree())
	if err != nil {
		t.Fatal(err)
	}

	ns := resolved.Namespaces[0]
	if ns.Capacity == nil || *ns.Capacity != (Capacity{Fixed: 32, Limit: 64}) {
		t.Fatalf("namespace capacity = %#v", ns.Capacity)
	}
	queue := ns.Queues[0]
	if queue.Capacity == nil || *queue.Capacity != (Capacity{Fixed: 32, Limit: 64}) {
		t.Fatalf("queue capacity = %#v", queue.Capacity)
	}
}

func TestResolveSiblingRemainder(t *testing.T) {
	tree := baseTree()
	tree.Namespaces = []Namespace{
		{
			Name:     "explicit",
			Capacity: capacity(8, 16),
			Queues:   []Queue{{Name: "q1", Capacity: capacity(8, 16)}},
		},
		{
			Name:   "remainder",
			Queues: []Queue{{Name: "q2"}},
		},
	}

	resolved, err := Resolve(tree)
	if err != nil {
		t.Fatal(err)
	}
	if got := *resolved.Namespaces[1].Capacity; got != (Capacity{Fixed: 24, Limit: 48}) {
		t.Fatalf("namespace remainder = %+v", got)
	}
	if got := *resolved.Namespaces[1].Queues[0].Capacity; got != (Capacity{Fixed: 24, Limit: 48}) {
		t.Fatalf("queue remainder = %+v", got)
	}
}

func TestResolveAllExplicitMayLeaveHeadroom(t *testing.T) {
	tree := baseTree()
	tree.Namespaces[0].Capacity = capacity(8, 16)
	tree.Namespaces[0].Queues[0].Capacity = capacity(4, 8)

	resolved, err := Resolve(tree)
	if err != nil {
		t.Fatal(err)
	}
	if got := *resolved.Namespaces[0].Capacity; got != (Capacity{Fixed: 8, Limit: 16}) {
		t.Fatalf("explicit namespace changed: %+v", got)
	}
}

func TestResolveAllowsExplicitZeroCapacityChildren(t *testing.T) {
	tree := baseTree()
	tree.Namespaces = []Namespace{
		{
			Name:     "disabled",
			Capacity: capacity(0, 0),
			Queues:   []Queue{{Name: "disabled-queue", Capacity: capacity(0, 0)}},
		},
		{
			Name:   "remainder",
			Queues: []Queue{{Name: "remainder-queue"}},
		},
	}

	resolved, err := Resolve(tree)
	if err != nil {
		t.Fatal(err)
	}
	if got := *resolved.Namespaces[0].Capacity; got != (Capacity{}) {
		t.Fatalf("explicit zero namespace capacity = %+v, want zero", got)
	}
	if got := *resolved.Namespaces[0].Queues[0].Capacity; got != (Capacity{}) {
		t.Fatalf("explicit zero queue capacity = %+v, want zero", got)
	}
	if got := *resolved.Namespaces[1].Capacity; got != (Capacity{Fixed: 32, Limit: 64}) {
		t.Fatalf("remainder namespace capacity = %+v", got)
	}
}

func TestResolveAllowsArbitraryFiniteUsedCU(t *testing.T) {
	tree := baseTree()
	tree.Workspace.Used = 4.1
	tree.Namespaces[0].Used = 4.1
	tree.Namespaces[0].Queues[0].Used = 4.1

	if _, err := Resolve(tree); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
}

func TestResolveValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Tree)
		wantErr string
	}{
		{
			name: "HA requires cross-zone fixed CU",
			mutate: func(tree *Tree) {
				tree.Workspace.CrossZoneFixedCU = 0
				tree.Workspace.FixedCU = 32
			},
			wantErr: "cross-zone",
		},
		{
			name: "non-HA forbids cross-zone fixed CU",
			mutate: func(tree *Tree) {
				tree.Workspace.HA = false
			},
			wantErr: "non-HA",
		},
		{
			name: "two namespace remainders",
			mutate: func(tree *Tree) {
				tree.Namespaces = append(tree.Namespaces, Namespace{Name: "second", Queues: []Queue{{Name: "q2"}}})
			},
			wantErr: "at most one",
		},
		{
			name: "duplicate queue names",
			mutate: func(tree *Tree) {
				tree.Namespaces[0].Queues = append(tree.Namespaces[0].Queues, Queue{Name: "default-queue", Capacity: capacity(1, 1)})
			},
			wantErr: "duplicate",
		},
		{
			name: "fixed children exceed parent",
			mutate: func(tree *Tree) {
				tree.Namespaces[0].Capacity = capacity(34, 34)
			},
			wantErr: "fixed",
		},
		{
			name: "limit children exceed parent",
			mutate: func(tree *Tree) {
				tree.Namespaces[0].Capacity = capacity(1, 66)
			},
			wantErr: "limit",
		},
		{
			name: "zero remainder",
			mutate: func(tree *Tree) {
				tree.Namespaces = []Namespace{
					{Name: "all", Capacity: capacity(32, 64), Queues: []Queue{{Name: "q1", Capacity: capacity(32, 64)}}},
					{Name: "zero", Queues: []Queue{{Name: "q2"}}},
				}
			},
			wantErr: "greater than zero",
		},
		{
			name: "pre requires fixed",
			mutate: func(tree *Tree) {
				tree.Workspace = WorkspaceCapacity{Limit: 2}
			},
			wantErr: "PRE",
		},
		{
			name: "post forbids fixed",
			mutate: func(tree *Tree) {
				tree.ChargeType = "POST"
			},
			wantErr: "POST",
		},
		{
			name: "post requires limit",
			mutate: func(tree *Tree) {
				tree.ChargeType = "POST"
				tree.Workspace = WorkspaceCapacity{}
			},
			wantErr: "limit",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree := baseTree()
			tc.mutate(&tree)
			_, err := Resolve(tree)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Resolve() error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestResolveDoesNotMutateInput(t *testing.T) {
	tree := baseTree()
	if _, err := Resolve(tree); err != nil {
		t.Fatal(err)
	}
	if tree.Namespaces[0].Capacity != nil || tree.Namespaces[0].Queues[0].Capacity != nil {
		t.Fatal("Resolve mutated input tree")
	}
}

func TestResolvePostChildrenArePureElastic(t *testing.T) {
	t.Run("omitted children receive zero fixed and the full limit", func(t *testing.T) {
		tree := Tree{
			ChargeType: "POST",
			Workspace:  WorkspaceCapacity{Limit: 16},
			Namespaces: []Namespace{{Name: "default", Queues: []Queue{{Name: "default-queue"}}}},
		}
		resolved, err := Resolve(tree)
		if err != nil {
			t.Fatal(err)
		}
		want := Capacity{Fixed: 0, Limit: 16}
		if *resolved.Namespaces[0].Capacity != want || *resolved.Namespaces[0].Queues[0].Capacity != want {
			t.Fatalf("resolved children = %#v", resolved.Namespaces[0])
		}
	})

	t.Run("explicit fixed child capacity is rejected", func(t *testing.T) {
		tree := Tree{
			ChargeType: "POST",
			Workspace:  WorkspaceCapacity{Limit: 16},
			Namespaces: []Namespace{{
				Name:     "default",
				Capacity: capacity(8, 16),
				Queues:   []Queue{{Name: "default-queue", Capacity: capacity(8, 16)}},
			}},
		}
		if _, err := Resolve(tree); err == nil || !strings.Contains(err.Error(), "fixed CU") {
			t.Fatalf("Resolve() error = %v, want fixed CU rejection", err)
		}
	})
}

func TestValidateDesiredRequiresOnlyImplicitDefaultQueue(t *testing.T) {
	tree := authoritativeTree(4, Namespace{
		Name:     "keep",
		Capacity: capacity(4, 4),
		Queues: []Queue{
			{Name: "default-queue", Capacity: capacity(4, 4)},
			{Name: "custom", Capacity: capacity(0, 0)},
		},
	})

	err := ValidateDesired(tree)
	if err == nil || !strings.Contains(err.Error(), `namespace "keep" must declare only implicit queue "default-queue"`) {
		t.Fatalf("ValidateDesired() error = %v", err)
	}
}
