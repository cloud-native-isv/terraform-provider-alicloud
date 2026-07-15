package flinkcapacity

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type API interface {
	ReadTree(context.Context, string) (Tree, error)
	ApplyStep(context.Context, string, Step) (Operation, error)
}

type Operation struct {
	RequestID string
	OrderID   string
}

type Reconciler struct {
	API              API
	PollInterval     time.Duration
	FinalReadTimeout time.Duration
	Now              func() time.Time
	Sleep            func(context.Context, time.Duration) error
}

type ReconcileError struct {
	Cause          error
	CompletedSteps int
	Step           *Step
	Operation      Operation
	FinalReadError error
}

func (e *ReconcileError) Error() string {
	message := fmt.Sprintf("capacity reconciliation failed after %d completed step(s)", e.CompletedSteps)
	if e.Step != nil {
		message += fmt.Sprintf(" at %s", describeStep(*e.Step))
	}
	if e.Operation.RequestID != "" {
		message += fmt.Sprintf(" (request_id=%s", e.Operation.RequestID)
		if e.Operation.OrderID != "" {
			message += fmt.Sprintf(", order_id=%s", e.Operation.OrderID)
		}
		message += ")"
	}
	message += ": " + e.Cause.Error()
	if e.FinalReadError != nil {
		message += "; final read failed: " + e.FinalReadError.Error()
	}
	return message
}

func (e *ReconcileError) Unwrap() error {
	return e.Cause
}

func (r Reconciler) Reconcile(ctx context.Context, instanceID string, desired Tree) (Tree, error) {
	if r.API == nil {
		return Tree{}, fmt.Errorf("capacity API must not be nil")
	}
	var current Tree
	var steps []Step
	var err error
	for {
		current, err = r.readTree(ctx, instanceID)
		if err != nil {
			return r.fail(instanceID, current, 0, nil, Operation{}, fmt.Errorf("read capacity tree before reconciliation: %w", err))
		}
		if desired.ChargeType == "" {
			desired.ChargeType = current.ChargeType
		}
		steps, err = Plan(current, desired)
		if err == nil {
			break
		}
		if !errorIsRetryable(err) {
			return current, err
		}
		if err := r.sleep(ctx); err != nil {
			return r.fail(instanceID, current, 0, nil, Operation{}, err)
		}
	}

	completed := 0
	for i := range steps {
		step := steps[i]
		current, err = r.readTree(ctx, instanceID)
		if err != nil {
			return r.fail(instanceID, current, completed, &step, Operation{}, fmt.Errorf("read before step: %w", err))
		}
		if stepConverged(current, step) {
			completed++
			continue
		}

		operation, applyErr := r.API.ApplyStep(ctx, instanceID, step)
		if applyErr != nil {
			if errorIsAmbiguous(applyErr) {
				verified, readErr := r.readTree(ctx, instanceID)
				if readErr == nil {
					current = verified
					if stepConverged(current, step) {
						completed++
						continue
					}
				}
			}
			return r.fail(instanceID, current, completed, &step, operation, applyErr)
		}

		for {
			current, err = r.readTree(ctx, instanceID)
			if err != nil {
				return r.fail(instanceID, current, completed, &step, operation, fmt.Errorf("read after step: %w", err))
			}
			if stepConverged(current, step) {
				completed++
				break
			}
			if err := r.sleep(ctx); err != nil {
				return r.fail(instanceID, current, completed, &step, operation, err)
			}
		}
	}
	return current, nil
}

func (r Reconciler) readTree(ctx context.Context, instanceID string) (Tree, error) {
	for {
		tree, err := r.API.ReadTree(ctx, instanceID)
		if err == nil || !errorIsRetryable(err) {
			return tree, err
		}
		if err := r.sleep(ctx); err != nil {
			return tree, err
		}
	}
}

func (r Reconciler) sleep(ctx context.Context) error {
	interval := r.PollInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if r.Sleep != nil {
		return r.Sleep(ctx, interval)
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (r Reconciler) fail(instanceID string, current Tree, completed int, step *Step, operation Operation, cause error) (Tree, error) {
	timeout := r.FinalReadTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	finalContext, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	finalTree, finalReadErr := r.API.ReadTree(finalContext, instanceID)
	if finalReadErr == nil {
		current = finalTree
	}
	return current, &ReconcileError{
		Cause:          cause,
		CompletedSteps: completed,
		Step:           step,
		Operation:      operation,
		FinalReadError: finalReadErr,
	}
}

func stepConverged(tree Tree, step Step) bool {
	resolved, err := Resolve(tree)
	if err != nil {
		return false
	}
	switch step.Action {
	case ModifyWorkspaceFixed:
		return resolved.Workspace.FixedCU == step.To.FixedCU && resolved.Workspace.CrossZoneFixedCU == step.To.CrossZoneFixedCU
	case ModifyWorkspacePostpaid:
		return resolved.Workspace.Limit == step.To.Limit
	case EnableWorkspaceElastic, ModifyWorkspaceElastic:
		return resolved.Workspace.AsCapacity().Elastic() == step.To.AsCapacity().Elastic()
	case ModifyNamespace:
		namespace := namespaceByName(resolved, step.Ref.Namespace)
		return namespace != nil && capacityAllocation(*namespace.Capacity) == step.To
	case ModifyQueue:
		namespace := namespaceByName(resolved, step.Ref.Namespace)
		if namespace == nil {
			return false
		}
		queue := queueByName(*namespace, step.Ref.Queue)
		return queue != nil && capacityAllocation(*queue.Capacity) == step.To
	default:
		return false
	}
}

func errorIsAmbiguous(err error) bool {
	var ambiguous interface{ Ambiguous() bool }
	return errors.As(err, &ambiguous) && ambiguous.Ambiguous()
}

func errorIsRetryable(err error) bool {
	var retryable interface{ Retryable() bool }
	return errors.As(err, &retryable) && retryable.Retryable()
}

func describeStep(step Step) string {
	if step.Ref.Queue != "" {
		return fmt.Sprintf("%s for queue %q/%q", step.Action, step.Ref.Namespace, step.Ref.Queue)
	}
	if step.Ref.Namespace != "" {
		return fmt.Sprintf("%s for namespace %q", step.Action, step.Ref.Namespace)
	}
	return string(step.Action)
}
