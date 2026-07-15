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
	desired := plannerTree(8, 16, 8, 8, 8, 8)
	api := &fakeCapacityAPI{tree: actual}
	api.applyFn = func(f *fakeCapacityAPI, step Step) (Operation, error) {
		applyCandidate(&f.tree, candidate{step: step})
		return Operation{RequestID: "ambiguous-request"}, ambiguousTestError{}
	}

	got, err := testReconciler(api).Reconcile(context.Background(), "f-test", desired)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.writes) != 1 {
		t.Fatalf("writes = %d, want 1", len(api.writes))
	}
	if !sameCapacityTree(got, desired) {
		t.Fatalf("final tree = %#v", got)
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
	desired := plannerTree(8, 16, 8, 8, 8, 8)
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

	_, err := testReconciler(api).Reconcile(context.Background(), "f-test", actual)
	if err == nil || !strings.Contains(err.Error(), "undeclared") {
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
