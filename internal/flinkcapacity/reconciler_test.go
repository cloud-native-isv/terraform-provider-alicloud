package flinkcapacity

import (
	"context"
	"errors"
	"strconv"
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

func TestReconcilerNetGrowthRenameStaysWithinEntryPlusOneAcrossCreateRestart(t *testing.T) {
	for _, sources := range []int{1, 2} {
		for _, ha := range []bool{false, true} {
			t.Run("sources="+strconv.Itoa(sources)+"/ha="+strconv.FormatBool(ha), func(t *testing.T) {
				actual, desired := authoritativeNetGrowthRenameFixture(sources, ha)
				ceiling := maxCU(actual.Workspace.Limit, desired.Workspace.Limit) + minimumNamespaceCU
				api := &fakeCapacityAPI{tree: cloneTree(actual)}
				interrupted := false
				api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
					applyCandidate(&f.tree, candidate{step: step})
					if !interrupted && step.Action == CreateNamespace {
						interrupted = true
						return Operation{RequestID: "create-written"}, errors.New("injected process restart after Create")
					}
					return Operation{RequestID: "request"}, nil
				}

				_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
				if err == nil || !interrupted || len(api.writes) != 1 || api.writes[0].Action != CreateNamespace || api.writes[0].Ref.Namespace != "new-00" {
					t.Fatalf("first reconcile error=%v interrupted=%t writes=%#v, want one applied replacement Create", err, interrupted, api.writes)
				}

				writesBeforeRestart := len(api.writes)
				api.applyFn = nil
				got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
				if err != nil {
					t.Fatal(err)
				}
				resumed := api.writes[writesBeforeRestart:]
				if len(resumed) == 0 || resumed[0].Action == DeleteNamespace {
					t.Fatalf("first post-restart write=%#v, want source preserved while replacement provenance remains ambiguous; all writes=%#v", resumed, api.writes)
				}
				usedBuffer := false
				for _, step := range api.writes {
					if isWorkspaceStep(step) && step.To.Limit > ceiling {
						t.Fatalf("net-growth rename exceeded operation-entry C0+1 %v after restart: %#v", ceiling.Float64(), api.writes)
					}
					if isWorkspaceStep(step) && step.To.Limit > maxCU(actual.Workspace.Limit, desired.Workspace.Limit) {
						usedBuffer = true
					}
				}
				if !usedBuffer {
					t.Fatalf("writes=%#v, want approved conservative C0+1 buffer after provenance-ambiguous restart", api.writes)
				}
				if !sameCapacityTree(got, desired) {
					t.Fatalf("final tree=%#v, want desired; writes=%#v", got, api.writes)
				}
			})
		}
	}
}

func TestReconcilerUnknownImportedPeakReusesHeadroomOrFailsClosed(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run("no-headroom/ha="+strconv.FormatBool(ha), func(t *testing.T) {
			actual, desired := authoritativeUnknownPeakFixture(ha, false)
			api := &fakeCapacityAPI{tree: actual}

			got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
			if err == nil || !strings.Contains(err.Error(), "no safe capacity transition") {
				t.Fatalf("tree=%#v error=%v, want fail-closed at unknown imported peak", got, err)
			}
			if len(api.writes) != 0 {
				t.Fatalf("writes=%#v, want no Delete guess and no second temporary CU", api.writes)
			}
		})

		t.Run("visible-headroom/ha="+strconv.FormatBool(ha), func(t *testing.T) {
			actual, desired := authoritativeUnknownPeakFixture(ha, true)
			api := &fakeCapacityAPI{tree: actual}

			got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
			if err != nil {
				t.Fatal(err)
			}
			if len(api.writes) == 0 || api.writes[0].Action != CreateNamespace {
				t.Fatalf("writes=%#v, want visible imported headroom reused for Create", api.writes)
			}
			for _, step := range api.writes {
				if isWorkspaceStep(step) && step.To.Limit > actual.Workspace.Limit {
					t.Fatalf("visible imported peak was stacked: actual=%v writes=%#v", actual.Workspace.Limit.Float64(), api.writes)
				}
			}
			if !sameCapacityTree(got, desired) {
				t.Fatalf("final tree=%#v, want desired; writes=%#v", got, api.writes)
			}
		})
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

func TestReconcilerAuthoritativePlannerFailuresWriteNothing(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		actual := authoritativeCompositionTree(false, 8, 12, 8)
		desired := authoritativeCompositionTree(false, 4, 12, 4)
		api := &fakeCapacityAPI{tree: actual}
		ctx := &cancelAfterChecksContext{Context: context.Background(), remaining: 4}

		_, err := testReconciler(api).ReconcileAuthoritative(ctx, "f-test", desired)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context cancellation", err)
		}
		if len(api.writes) != 0 {
			t.Fatalf("writes = %#v, want no writes after planning cancellation", api.writes)
		}
	})

	t.Run("hard planner budget", func(t *testing.T) {
		actual := authoritativeCompositionTree(false, schemaMaximumCUHalfUnits, schemaMaximumCUHalfUnits, schemaMaximumCUHalfUnits)
		desired := authoritativeCompositionTree(false, minimumNamespaceCU, schemaMaximumCUHalfUnits, minimumNamespaceCU)
		api := &fakeCapacityAPI{tree: actual}

		_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
		var budgetErr *PlannerBudgetError
		if !errors.As(err, &budgetErr) {
			t.Fatalf("error = %T %v, want PlannerBudgetError", err, err)
		}
		if len(api.writes) != 0 {
			t.Fatalf("writes = %#v, want no writes after planner budget exhaustion", api.writes)
		}
	})

	t.Run("schema maximum rename has no serializable buffer", func(t *testing.T) {
		for _, ha := range []bool{false, true} {
			t.Run("ha="+strconv.FormatBool(ha), func(t *testing.T) {
				actual := authoritativeCompositionTree(ha, schemaMaximumCUHalfUnits, schemaMaximumCUHalfUnits, schemaMaximumCUHalfUnits)
				actual.Namespaces[0].Name = "old"
				desired := cloneTree(actual)
				desired.Namespaces[0].Name = "new"
				api := &fakeCapacityAPI{tree: actual}

				_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
				if err == nil || !strings.Contains(err.Error(), "no safe capacity transition") {
					t.Fatalf("error=%v, want terminal no-safe error", err)
				}
				if len(api.writes) != 0 {
					t.Fatalf("writes=%#v, want zero writes for an unserializable topology buffer", api.writes)
				}
			})
		}
	})
}

func TestReconcilerAuthoritativeRestartDoesNotStackTemporaryCapacity(t *testing.T) {
	tests := []struct {
		name      string
		actual    Tree
		desired   Tree
		interrupt func(Step) bool
	}{
		{
			name:    "after topology buffer",
			actual:  authoritativeTree(4, authoritativeNamespace("old", 4, 4, 0)),
			desired: authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0)),
			interrupt: func(step Step) bool {
				return step.Action == ModifyWorkspaceFixed && step.Temporary
			},
		},
		{
			name:    "after create",
			actual:  authoritativeTree(4, authoritativeNamespace("old", 4, 4, 0)),
			desired: authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0)),
			interrupt: func(step Step) bool {
				return step.Action == CreateNamespace
			},
		},
		{
			name:    "after delete",
			actual:  authoritativeTree(4, authoritativeNamespace("old", 4, 4, 0)),
			desired: authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0)),
			interrupt: func(step Step) bool {
				return step.Action == DeleteNamespace
			},
		},
		{
			name:    "after composition bridge",
			actual:  authoritativeCompositionTree(false, 8, 12, 8),
			desired: authoritativeCompositionTree(false, 4, 12, 4),
			interrupt: func(step Step) bool {
				return isWorkspaceStep(step) && step.Temporary
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entryCeiling := maxCU(tc.actual.Workspace.Limit, tc.desired.Workspace.Limit) + minimumNamespaceCU
			api := &fakeCapacityAPI{tree: cloneTree(tc.actual)}
			interrupted := false
			api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
				applyCandidate(&f.tree, candidate{step: step})
				if !interrupted && tc.interrupt(step) {
					interrupted = true
					return Operation{RequestID: "interrupted-after-write"}, errors.New("injected interruption")
				}
				return Operation{RequestID: "request"}, nil
			}

			_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", tc.desired)
			if err == nil || !interrupted {
				t.Fatalf("first reconcile error=%v interrupted=%t writes=%#v", err, interrupted, api.writes)
			}
			peak := tc.actual.Workspace.Limit
			for _, step := range api.writes {
				if isWorkspaceStep(step) && step.To.Limit > peak {
					peak = step.To.Limit
				}
			}
			writesBeforeResume := len(api.writes)
			api.applyFn = nil

			got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", tc.desired)
			if err != nil {
				t.Fatal(err)
			}
			for _, step := range api.writes[writesBeforeResume:] {
				if isWorkspaceStep(step) && step.To.Limit > peak {
					t.Fatalf("resume stacked temporary capacity above %v: %#v", peak.Float64(), api.writes)
				}
			}
			for _, step := range api.writes {
				if isWorkspaceStep(step) && step.To.Limit > entryCeiling {
					t.Fatalf("failure/restart path exceeded operation-entry C0+1 %v: %#v", entryCeiling.Float64(), api.writes)
				}
			}
			if !sameCapacityTree(got, tc.desired) {
				t.Fatalf("resumed tree = %#v, want desired; writes=%#v", got, api.writes)
			}
		})
	}
}

func TestReconcilerTopologyBufferErrorsDoNotStackCapacity(t *testing.T) {
	for _, failure := range []struct {
		name      string
		applyStep bool
		ambiguous bool
		err       error
	}{
		{name: "ordinary before write", err: errors.New("injected terminal write rejection")},
		{name: "ordinary after write", applyStep: true, err: errors.New("injected process loss after write")},
		{name: "ambiguous after write", applyStep: true, ambiguous: true, err: ambiguousTestError{}},
	} {
		t.Run(failure.name, func(t *testing.T) {
			actual := authoritativeTree(4, authoritativeNamespace("old", 4, 4, 0))
			desired := authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0))
			ceiling := actual.Workspace.Limit + minimumNamespaceCU
			api := &fakeCapacityAPI{tree: cloneTree(actual)}
			injected := false
			api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
				if !injected && step.Action == ModifyWorkspaceFixed && step.Temporary {
					injected = true
					if failure.applyStep {
						applyCandidate(&f.tree, candidate{step: step})
					}
					return Operation{RequestID: "buffer-request"}, failure.err
				}
				applyCandidate(&f.tree, candidate{step: step})
				return Operation{RequestID: "request"}, nil
			}

			got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
			if !injected {
				t.Fatalf("buffer fault was not injected; writes=%#v", api.writes)
			}
			if failure.ambiguous {
				if err != nil {
					t.Fatalf("ambiguous applied buffer did not recover from Read/replan: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("ordinary buffer failure returned success; writes=%#v", api.writes)
				}
				api.applyFn = nil
				got, err = testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
				if err != nil {
					t.Fatalf("restart after ordinary buffer failure did not recover: %v", err)
				}
			}
			for _, step := range api.writes {
				if isWorkspaceStep(step) && step.To.Limit > ceiling {
					t.Fatalf("buffer error path stacked above C0+1 %v: %#v", ceiling.Float64(), api.writes)
				}
			}
			if !sameCapacityTree(got, desired) {
				t.Fatalf("final tree=%#v, want desired; writes=%#v", got, api.writes)
			}
		})
	}
}

func TestReconcilerPartialHeadroomTopologyBufferRecoversWithoutElasticOrStacking(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run("resume/ha="+strconv.FormatBool(ha), func(t *testing.T) {
			actual, desired := authoritativePartialHeadroomRenameFixture(ha)
			api := &fakeCapacityAPI{tree: cloneTree(actual)}
			interrupted := false
			api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
				applyCandidate(&f.tree, candidate{step: step})
				if !interrupted && isWorkspaceStep(step) && step.To.Limit > actual.Workspace.Limit {
					interrupted = true
					return Operation{RequestID: "buffer-written"}, errors.New("injected interruption after topology buffer")
				}
				return Operation{RequestID: "request"}, nil
			}

			_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
			if err == nil || !interrupted {
				t.Fatalf("first reconcile error=%v interrupted=%t writes=%#v", err, interrupted, api.writes)
			}
			first := api.writes[0]
			if first.Action != ModifyWorkspaceFixed || first.To.AsCapacity().Elastic() != first.From.AsCapacity().Elastic() {
				t.Fatalf("interrupted first write = %#v, want pure fixed topology buffer", first)
			}
			peak := first.To.Limit
			writesBeforeResume := len(api.writes)
			api.applyFn = nil

			got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
			if err != nil {
				t.Fatal(err)
			}
			for _, step := range api.writes[writesBeforeResume:] {
				if isWorkspaceStep(step) && step.To.Limit > peak {
					t.Fatalf("restart stacked topology capacity above %v: %#v", peak.Float64(), api.writes)
				}
			}
			if !sameCapacityTree(got, desired) {
				t.Fatalf("resumed tree = %#v, want desired; writes=%#v", got, api.writes)
			}
		})

		t.Run("rollback-to-zero-elastic/ha="+strconv.FormatBool(ha), func(t *testing.T) {
			actual, desired := authoritativePartialHeadroomRenameFixture(ha)
			rollback := authoritativeCompositionTree(ha, 10, 10, 10)
			rollback.Namespaces[0].Name = "old"
			api := &fakeCapacityAPI{tree: cloneTree(actual)}
			api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
				applyCandidate(&f.tree, candidate{step: step})
				return Operation{RequestID: "buffer-written"}, errors.New("injected interruption after first write")
			}

			_, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", desired)
			if err == nil || len(api.writes) != 1 {
				t.Fatalf("first reconcile error=%v writes=%#v, want one applied buffer then failure", err, api.writes)
			}
			buffer := api.writes[0]
			if buffer.Action != ModifyWorkspaceFixed || buffer.To.AsCapacity().Elastic() != 0 {
				t.Fatalf("first write = %#v, want rollback-safe fixed buffer with E=0", buffer)
			}

			api.applyFn = nil
			writesBeforeRollback := len(api.writes)
			got, err := testReconciler(api).ReconcileAuthoritative(context.Background(), "f-test", rollback)
			if err != nil {
				t.Fatalf("rollback to E=0 failed: %v; writes=%#v", err, api.writes)
			}
			for _, step := range api.writes[writesBeforeRollback:] {
				if step.Action == EnableWorkspaceElastic || step.Action == ModifyWorkspaceElastic {
					t.Fatalf("rollback used elastic write: %#v", api.writes)
				}
			}
			if !sameCapacityTree(got, rollback) {
				t.Fatalf("rollback tree = %#v, want original E=0 tree; writes=%#v", got, api.writes)
			}
		})
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
