package flinkcapacity

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeCapacityAPI struct {
	tree      Tree
	reads     int
	writes    []Step
	deadlines []time.Time
	readFn    func(*fakeCapacityAPI) (Tree, error)
	applyFn   func(*fakeCapacityAPI, Step) (Operation, error)
}

func (f *fakeCapacityAPI) ReadTree(ctx context.Context, _ string) (Tree, error) {
	f.reads++
	if deadline, ok := ctx.Deadline(); ok {
		f.deadlines = append(f.deadlines, deadline)
	}
	if f.readFn != nil {
		return f.readFn(f)
	}
	return cloneTree(f.tree), nil
}

func (f *fakeCapacityAPI) ApplyStep(ctx context.Context, _ string, step Step) (Operation, error) {
	f.writes = append(f.writes, step)
	if deadline, ok := ctx.Deadline(); ok {
		f.deadlines = append(f.deadlines, deadline)
	}
	if f.applyFn != nil {
		return f.applyFn(f, step)
	}
	applyCandidate(&f.tree, candidate{step: step})
	return Operation{RequestID: "request", OrderID: "order"}, nil
}

type ambiguousTestError struct{}

func (ambiguousTestError) Error() string   { return "connection reset after request" }
func (ambiguousTestError) Ambiguous() bool { return true }

type retryableTestError struct{}

func (retryableTestError) Error() string   { return "not ready" }
func (retryableTestError) Retryable() bool { return true }

func testReconciler(api API) Reconciler {
	return Reconciler{
		API:          api,
		PollInterval: time.Millisecond,
		Sleep: func(context.Context, time.Duration) error {
			return nil
		},
	}
}

func TestReconcilerReplansFromTheObservedPostWriteTreeAndSharesDeadline(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	desired := plannerTree(16, 24, 16, 24, 16, 24)
	api := &fakeCapacityAPI{tree: actual}
	reconciler := testReconciler(api)
	deadline := time.Now().Add(time.Hour).Round(0)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	got, err := reconciler.Reconcile(ctx, "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("final tree = %#v", got)
	}
	if len(api.writes) == 0 || api.reads != len(api.writes)+1 {
		t.Fatalf("reads=%d writes=%d, want one initial read and one post-write read per step", api.reads, len(api.writes))
	}
	for _, gotDeadline := range api.deadlines {
		if !gotDeadline.Equal(deadline) {
			t.Fatalf("deadline changed: %v != %v", gotDeadline, deadline)
		}
	}
}

func TestReconcilerDoesNotReplayAmbiguousSuccess(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	desired := plannerTree(8, 16, 8, 16, 8, 16)
	api := &fakeCapacityAPI{tree: actual}
	api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
		applyCandidate(&f.tree, candidate{step: step})
		return Operation{RequestID: "ambiguous-request"}, ambiguousTestError{}
	}

	got, err := testReconciler(api).Reconcile(context.Background(), "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.writes) != 3 {
		t.Fatalf("writes = %d, want one non-replayed write per planned step", len(api.writes))
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("final tree = %#v", got)
	}
}

func TestReconcilerPollsAmbiguousWritesUntilASecondReadConverges(t *testing.T) {
	tests := []struct {
		name          string
		actual        Tree
		desired       Tree
		wantAction    Action
		wantRefName   string
		authoritative bool
	}{
		{
			name:       "capacity",
			actual:     plannerTree(8, 8, 8, 8, 8, 8),
			desired:    plannerTree(8, 16, 8, 16, 8, 16),
			wantAction: EnableWorkspaceElastic,
		},
		{
			name: "create namespace",
			actual: authoritativeTree(6,
				authoritativeNamespace("keep", 2, 2, 0),
			),
			desired: authoritativeTree(6,
				authoritativeNamespace("keep", 2, 2, 0),
				authoritativeNamespace("new", 4, 4, 0),
			),
			wantAction:    CreateNamespace,
			wantRefName:   "new",
			authoritative: true,
		},
		{
			name: "delete namespace",
			actual: authoritativeTree(4,
				authoritativeNamespace("keep", 2, 2, 0),
				authoritativeNamespace("drop", 2, 2, 0),
			),
			desired:       authoritativeTree(2, authoritativeNamespace("keep", 2, 2, 0)),
			wantAction:    DeleteNamespace,
			wantRefName:   "drop",
			authoritative: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := &fakeCapacityAPI{tree: cloneTree(tc.actual)}
			var pending *Step
			ambiguousPolls := 0
			api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
				if len(f.writes) == 1 {
					copy := step
					pending = &copy
					return Operation{RequestID: "ambiguous-request"}, ambiguousTestError{}
				}
				applyCandidate(&f.tree, candidate{step: step})
				return Operation{RequestID: "request"}, nil
			}
			api.readFn = func(f *fakeCapacityAPI) (Tree, error) {
				if pending != nil {
					ambiguousPolls++
					if ambiguousPolls == 2 {
						applyCandidate(&f.tree, candidate{step: *pending})
						pending = nil
					}
				}
				return cloneTree(f.tree), nil
			}

			reconciler := testReconciler(api)
			var got Tree
			var err error
			if tc.authoritative {
				got, err = reconciler.ReconcileAuthoritative(context.Background(), "f-test", tc.desired)
			} else {
				got, err = reconciler.Reconcile(context.Background(), "f-test", tc.desired)
			}
			if err != nil {
				t.Fatal(err)
			}
			if ambiguousPolls != 2 {
				t.Fatalf("ambiguous post-write reads = %d, want 2", ambiguousPolls)
			}
			first := api.writes[0]
			if first.Action != tc.wantAction || first.Ref.Namespace != tc.wantRefName {
				t.Fatalf("first write = %#v", first)
			}
			firstWrites := 0
			for _, step := range api.writes {
				if step.Action == first.Action && step.Ref == first.Ref && step.To == first.To {
					firstWrites++
				}
			}
			if firstWrites != 1 {
				t.Fatalf("ambiguous step replayed %d times: %#v", firstWrites, api.writes)
			}
			if !sameCapacityTree(got, tc.desired) {
				t.Fatalf("final tree = %#v", got)
			}
		})
	}
}

func TestReconcilerReturnsPartialTreeOnLaterFailure(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	desired := plannerTree(16, 24, 16, 24, 16, 24)
	api := &fakeCapacityAPI{tree: actual}
	api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
		if len(f.writes) == 2 {
			return Operation{RequestID: "failed-request", OrderID: "failed-order"}, errors.New("order failed")
		}
		applyCandidate(&f.tree, candidate{step: step})
		return Operation{RequestID: "successful-request"}, nil
	}

	got, err := testReconciler(api).Reconcile(context.Background(), "f-test", desired)
	if err == nil {
		t.Fatal("expected reconcile failure")
	}
	var reconcileErr *ReconcileError
	if !errors.As(err, &reconcileErr) {
		t.Fatalf("error type = %T, want *ReconcileError", err)
	}
	if reconcileErr.CompletedSteps != 1 || reconcileErr.Operation.RequestID != "failed-request" || reconcileErr.Operation.OrderID != "failed-order" {
		t.Fatalf("diagnostic = %#v", reconcileErr)
	}
	if !stepConverged(got, api.writes[0]) {
		t.Fatalf("returned tree lost the first successful step: %#v", got)
	}
}

func TestReconcilerFinalReadOnTimeout(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	desired := plannerTree(8, 16, 8, 16, 8, 16)
	api := &fakeCapacityAPI{tree: actual}
	api.applyFn = func(_ *fakeCapacityAPI, _ Step) (Operation, error) {
		return Operation{RequestID: "slow-request"}, nil
	}
	reconciler := testReconciler(api)
	reconciler.Sleep = func(context.Context, time.Duration) error {
		return context.DeadlineExceeded
	}

	_, err := reconciler.Reconcile(context.Background(), "f-test", desired)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
	if api.reads < 3 {
		t.Fatalf("reads = %d, want pre-write, post-write, and final reads", api.reads)
	}
}

func TestReconcilerInheritsObservedChargeType(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	desired := cloneTree(actual)
	desired.ChargeType = ""
	api := &fakeCapacityAPI{tree: actual}

	if _, err := testReconciler(api).Reconcile(context.Background(), "f-test", desired); err != nil {
		t.Fatal(err)
	}
	if len(api.writes) != 0 {
		t.Fatalf("writes = %d, want 0", len(api.writes))
	}
}

func TestReconcilerWaitsForRetryableInitialRead(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	api := &fakeCapacityAPI{tree: actual}
	api.readFn = func(f *fakeCapacityAPI) (Tree, error) {
		if f.reads < 3 {
			return Tree{}, retryableTestError{}
		}
		return cloneTree(f.tree), nil
	}

	if _, err := testReconciler(api).Reconcile(context.Background(), "f-test", actual); err != nil {
		t.Fatal(err)
	}
	if api.reads != 3 {
		t.Fatalf("reads = %d, want 3", api.reads)
	}
}

func TestReconcilerWaitsForDeclaredTopology(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	missing := cloneTree(actual)
	missing.Namespaces[0].Queues = nil
	api := &fakeCapacityAPI{tree: missing}
	api.readFn = func(f *fakeCapacityAPI) (Tree, error) {
		if f.reads == 1 {
			return cloneTree(missing), nil
		}
		f.tree = cloneTree(actual)
		return cloneTree(actual), nil
	}

	if _, err := testReconciler(api).Reconcile(context.Background(), "f-test", actual); err != nil {
		t.Fatal(err)
	}
	if api.reads != 2 {
		t.Fatalf("reads = %d, want 2", api.reads)
	}
}

func TestReconcilerFailsClosedForUndeclaredTopology(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	extra := cloneTree(actual)
	extra.Namespaces[0].Queues = append(extra.Namespaces[0].Queues, Queue{
		Name:     "unexpected",
		Capacity: &Capacity{Fixed: 0, Limit: 0},
	})
	api := &fakeCapacityAPI{tree: extra}

	_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", actual)
	if err == nil || !strings.Contains(err.Error(), "unsupported queue") {
		t.Fatalf("error = %v", err)
	}
	if api.reads != 1 {
		t.Fatalf("reads = %d, want no retry", api.reads)
	}
}

func TestReconcilerReplansFromObservedPostWriteTree(t *testing.T) {
	actual := plannerTree(8, 8, 8, 8, 8, 8)
	desired := plannerTree(16, 16, 16, 16, 16, 16)
	api := &fakeCapacityAPI{tree: actual}
	api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
		next := cloneTree(f.tree)
		applyCandidate(&next, candidate{step: step})
		if len(f.writes) == 1 {
			// The authoritative post-write response may include a concurrent
			// control-plane change to a child. The next step must be planned
			// from that response rather than the pre-write plan.
			next.Namespaces[0].Capacity = &Capacity{Fixed: 12, Limit: 12}
		}
		if _, err := Resolve(next); err != nil {
			return Operation{}, errors.New("reconciler attempted an unsafe stale step: " + err.Error())
		}
		f.tree = next
		return Operation{RequestID: "request"}, nil
	}

	if _, err := testReconciler(api).Reconcile(context.Background(), "f-test", desired); err != nil {
		t.Fatal(err)
	}
	foundReplannedNamespace := false
	for _, step := range api.writes {
		if step.Action == ModifyNamespace && step.From.FixedCU == 12 {
			foundReplannedNamespace = true
		}
	}
	if !foundReplannedNamespace {
		t.Fatalf("steps did not replan from the observed namespace capacity: %#v", api.writes)
	}
}

func TestReconcilerReplansAfterNamespaceTopologySteps(t *testing.T) {
	actual := authoritativeTree(4, authoritativeNamespace("old", 4, 4, 0))
	desired := authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0))
	api := &fakeCapacityAPI{tree: actual}

	got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("final tree = %#v", got)
	}
	if api.reads != len(api.writes)+1 {
		t.Fatalf("reads=%d writes=%d, want one initial read and one post-write read per step", api.reads, len(api.writes))
	}
}

func TestReconcilerReusesOneCUTemporaryBufferAcrossMultipleRenames(t *testing.T) {
	actual := authoritativeTree(4,
		authoritativeNamespace("old-a", 2, 2, 0),
		authoritativeNamespace("old-b", 2, 2, 0),
	)
	desired := authoritativeTree(4,
		authoritativeNamespace("new-a", 2, 2, 0),
		authoritativeNamespace("new-b", 2, 2, 0),
	)
	api := &fakeCapacityAPI{tree: actual}

	got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	var topology []string
	bufferExpansions := 0
	peak := CU(0)
	for _, step := range api.writes {
		if step.Action == CreateNamespace || step.Action == DeleteNamespace {
			topology = append(topology, string(step.Action)+":"+step.Ref.Namespace)
		}
		if step.Action == ModifyWorkspaceFixed {
			if step.To.TotalFixed() > peak {
				peak = step.To.TotalFixed()
			}
			if step.To.TotalFixed() > 4 {
				bufferExpansions++
			}
		}
	}
	want := []string{
		"create_namespace:new-a",
		"delete_namespace:old-a",
		"create_namespace:new-b",
		"delete_namespace:old-b",
	}
	if strings.Join(topology, ",") != strings.Join(want, ",") {
		t.Fatalf("topology steps = %#v, want %#v; all writes = %#v", topology, want, api.writes)
	}
	if bufferExpansions != 1 || peak != 6 {
		t.Fatalf("buffer expansions=%d peak=%v, want one expansion to 6 half-CU; writes=%#v", bufferExpansions, peak, api.writes)
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("final tree = %#v", got)
	}
}

func TestReconcilerRenameAndShrinkBuffersFromMigrationOriginWithoutStacking(t *testing.T) {
	actual := authoritativeTree(6, authoritativeNamespace("old", 6, 6, 0))
	desired := authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0))
	api := &fakeCapacityAPI{tree: actual}

	got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	bufferExpansions := 0
	peak := CU(0)
	for _, step := range api.writes {
		if step.Action != ModifyWorkspaceFixed {
			continue
		}
		if step.To.TotalFixed() > peak {
			peak = step.To.TotalFixed()
		}
		if step.To.TotalFixed() > 6 {
			bufferExpansions++
		}
	}
	if bufferExpansions != 1 || peak != 8 {
		t.Fatalf("buffer expansions=%d peak=%v, want one origin-relative expansion to 8 half-CU; writes=%#v", bufferExpansions, peak, api.writes)
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("final tree = %#v", got)
	}
}

func TestReconcilerRefreshesWorkspaceUsedAfterDeletingBusyNamespace(t *testing.T) {
	actual := authoritativeTree(6,
		authoritativeNamespace("keep", 2, 2, 0.5),
		authoritativeNamespace("drop", 4, 4, 2.5),
	)
	actual.Workspace.Used = 3
	desired := authoritativeTree(2, authoritativeNamespace("keep", 2, 2, 0))
	api := &fakeCapacityAPI{tree: actual}
	api.readFn = func(f *fakeCapacityAPI) (Tree, error) {
		if namespaceByName(f.tree, "drop") == nil {
			f.tree.Workspace.Used = 0.5
		}
		return cloneTree(f.tree), nil
	}

	got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.writes) < 2 || api.writes[0].Action != DeleteNamespace || api.writes[1].Action != ModifyWorkspaceFixed {
		t.Fatalf("writes = %#v, want delete barrier before workspace shrink", api.writes)
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("final tree = %#v", got)
	}
}

func TestReconcilerPreservesRenameSourceWhileCrossedAllocationsConverge(t *testing.T) {
	actual := authoritativeTree(8,
		authoritativeNamespace("a", 4, 8, 0),
		authoritativeNamespace("b", 2, 10, 0),
		authoritativeNamespace("old", 2, 2, 0),
	)
	actual.Workspace.Limit = 20
	desired := authoritativeTree(8,
		authoritativeNamespace("a", 2, 10, 0),
		authoritativeNamespace("b", 4, 8, 0),
		authoritativeNamespace("new", 2, 2, 0),
	)
	desired.Workspace.Limit = 20
	api := &fakeCapacityAPI{tree: actual}
	api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
		if step.Action == DeleteNamespace && step.Ref.Namespace == "old" && namespaceByName(f.tree, "new") == nil {
			return Operation{}, errors.New("rename source deleted before replacement existed")
		}
		applyCandidate(&f.tree, candidate{step: step})
		return Operation{RequestID: "request"}, nil
	}

	got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("final tree = %#v; writes=%#v", got, api.writes)
	}
}

func TestLegacyReconcilerAllowsMultipleQueuesAndRemainder(t *testing.T) {
	actual := Tree{
		ChargeType: "PRE",
		Workspace:  WorkspaceCapacity{FixedCU: 8, Limit: 16},
		Namespaces: []Namespace{{
			Name:     "legacy",
			Capacity: capacity(8, 16),
			Queues: []Queue{
				{Name: "explicit", Capacity: capacity(2, 4)},
				{Name: "remainder", Capacity: capacity(6, 12)},
			},
		}},
	}
	desired := Tree{
		ChargeType: "PRE",
		Workspace:  WorkspaceCapacity{FixedCU: 8, Limit: 16},
		Namespaces: []Namespace{{
			Name: "legacy",
			Queues: []Queue{
				{Name: "explicit", Capacity: capacity(2, 4)},
				{Name: "remainder"},
			},
		}},
	}
	api := &fakeCapacityAPI{tree: actual}

	got, err := testReconciler(api).Reconcile(context.Background(), "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.writes) != 0 {
		t.Fatalf("legacy no-op reconciliation wrote steps: %#v", api.writes)
	}
	if !legacySameCapacityTree(got, actual) {
		t.Fatalf("legacy tree changed: %#v", got)
	}
}

func TestReconcilerAuthoritativeRejectsPostpaidBeforeWrites(t *testing.T) {
	postpaid := postpaidAuthoritativeTree(8)
	prepaid := authoritativeTree(minimumNamespaceCU,
		authoritativeNamespace("default", minimumNamespaceCU, minimumNamespaceCU, 0),
	)

	for _, tc := range []struct {
		name    string
		actual  Tree
		desired Tree
	}{
		{name: "actual POST", actual: postpaid, desired: cloneTree(postpaid)},
		{name: "desired POST", actual: prepaid, desired: postpaid},
		{name: "empty desired inherits actual POST", actual: postpaid, desired: Tree{Namespaces: cloneTree(postpaid).Namespaces, Workspace: cloneTree(postpaid).Workspace}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := &fakeCapacityAPI{tree: tc.actual}

			_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", tc.desired)
			if err == nil || !strings.Contains(err.Error(), "PRE") {
				t.Fatalf("ReconcileAuthoritative() error = %v, want PRE-only error", err)
			}
			if len(api.writes) != 0 {
				t.Fatalf("ReconcileAuthoritative() wrote %#v before rejecting POST workspace", api.writes)
			}
		})
	}
}

func TestReconcilerAuthoritativeRejectsNamespaceBelowMinimumBeforeWrites(t *testing.T) {
	actual := authoritativeTree(minimumNamespaceCU,
		authoritativeNamespace("default", minimumNamespaceCU, minimumNamespaceCU, 0),
	)
	desired := authoritativeTree(minimumNamespaceCU-1,
		authoritativeNamespace("default", minimumNamespaceCU-1, minimumNamespaceCU-1, 0),
	)
	api := &fakeCapacityAPI{tree: actual}

	_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
	if err == nil || !strings.Contains(err.Error(), "at least 1") {
		t.Fatalf("ReconcileAuthoritative() error = %v, want one-CU namespace minimum", err)
	}
	if len(api.writes) != 0 {
		t.Fatalf("ReconcileAuthoritative() wrote %#v before rejecting namespace minimum", api.writes)
	}
}

func TestStepConvergedTopologyDoesNotRequireUnrelatedDeletedQueuesToResolve(t *testing.T) {
	tree := authoritativeTree(4,
		namespaceWithQueues(
			"drop", 2, 2,
			Queue{Name: "default-queue", Capacity: capacity(2, 2)},
			Queue{Name: "custom", Capacity: capacity(1, 1)},
		),
		authoritativeNamespace("new", 2, 2, 0),
	)

	if !stepConverged(tree, Step{Action: CreateNamespace, Ref: Ref{Namespace: "new"}}) {
		t.Fatal("created namespace must converge even when an undeclared namespace has custom queues")
	}
}
