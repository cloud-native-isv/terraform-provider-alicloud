package flinkcapacity

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

const schemaMaximumCUHalfUnits CU = 536_870_911 * 2

func authoritativeCompositionTree(ha bool, fixed, limit, namespaceFixed CU) Tree {
	namespace := authoritativeNamespace("default", namespaceFixed, limit, 0)
	namespace.CrossZone = ha
	workspace := WorkspaceCapacity{Limit: limit, HA: ha}
	if ha {
		workspace.CrossZoneFixedCU = fixed
	} else {
		workspace.FixedCU = fixed
	}
	return Tree{ChargeType: "PRE", Workspace: workspace, Namespaces: []Namespace{namespace}}
}

func applyPlanForTest(t *testing.T, actual Tree, steps []Step) Tree {
	t.Helper()
	state := cloneTree(actual)
	for i, step := range steps {
		if step.Action == DeleteNamespace && i != len(steps)-1 {
			t.Fatalf("delete at step %d is not the observation frontier: %#v", i, steps)
		}
		applyCandidate(&state, candidate{step: step})
		if err := validatePlanningState(state); err != nil {
			t.Fatalf("step %d produced unsafe state: %v; step=%#v plan=%#v", i, err, step, steps)
		}
	}
	return state
}

func TestPlanAuthoritativeCompositionBothDirectionsSingleAndHA(t *testing.T) {
	tests := []struct {
		name    string
		actual  func(bool) Tree
		desired func(bool) Tree
	}{
		{
			name:    "fixed to elastic",
			actual:  func(ha bool) Tree { return authoritativeCompositionTree(ha, 8, 12, 8) },
			desired: func(ha bool) Tree { return authoritativeCompositionTree(ha, 4, 12, 4) },
		},
		{
			name:    "elastic to fixed",
			actual:  func(ha bool) Tree { return authoritativeCompositionTree(ha, 4, 12, 4) },
			desired: func(ha bool) Tree { return authoritativeCompositionTree(ha, 8, 12, 8) },
		},
	}
	for _, tc := range tests {
		for _, ha := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/ha=%t", tc.name, ha), func(t *testing.T) {
				actual := tc.actual(ha)
				desired := tc.desired(ha)
				steps, err := PlanAuthoritative(actual, desired)
				if err != nil {
					t.Fatal(err)
				}
				again, err := PlanAuthoritative(actual, desired)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(steps, again) {
					t.Fatalf("plan is nondeterministic: %#v != %#v", steps, again)
				}
				state := applyPlanForTest(t, actual, steps)
				if !sameCapacityTree(state, desired) {
					t.Fatalf("plan did not converge: %#v", steps)
				}
				peak := actual.Workspace.Limit
				temporaryWorkspaceSteps := 0
				for _, step := range steps {
					if isWorkspaceStep(step) {
						if step.To.Limit > peak {
							peak = step.To.Limit
						}
						if ha && step.To.FixedCU != 0 {
							t.Fatalf("HA plan modified primary fixed pool: %#v", step)
						}
						if step.Temporary {
							temporaryWorkspaceSteps++
						}
					}
				}
				if peak != 14 || temporaryWorkspaceSteps == 0 {
					t.Fatalf("peak=%v temporary_steps=%d, want exactly one-CU bridge entitlement; steps=%#v", peak.Float64(), temporaryWorkspaceSteps, steps)
				}
			})
		}
	}
}

func TestPlanAuthoritativeExhaustsEndpointScheduleBeforeBridge(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(fmt.Sprintf("ha=%t", ha), func(t *testing.T) {
			// F2/E4/L6 -> F4/E3/L7 has a complete endpoint-bounded path:
			// fixed and elastic can alternate using L7 itself as the bridge.
			actual := authoritativeCompositionTree(ha, 4, 12, 4)
			desired := authoritativeCompositionTree(ha, 8, 14, 8)
			steps, err := PlanAuthoritative(actual, desired)
			if err != nil {
				t.Fatal(err)
			}
			for _, step := range steps {
				if isWorkspaceStep(step) && step.To.Limit > desired.Workspace.Limit {
					t.Fatalf("endpoint-bounded plan unnecessarily exceeded L7: %#v", steps)
				}
			}
			if state := applyPlanForTest(t, actual, steps); !sameCapacityTree(state, desired) {
				t.Fatalf("plan did not converge: %#v", steps)
			}
		})
	}
}

func TestPlanAuthoritativeRenameCompositionStopsAtDeleteFrontier(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(fmt.Sprintf("ha=%t", ha), func(t *testing.T) {
			actual := authoritativeCompositionTree(ha, 8, 10, 8)
			actual.Namespaces[0].Name = "old"
			actual.Workspace.Used = 4
			desired := authoritativeCompositionTree(ha, 4, 12, 4)
			desired.Namespaces[0].Name = "new"

			steps, err := PlanAuthoritative(actual, desired)
			if err != nil {
				t.Fatal(err)
			}
			if len(steps) < 3 || steps[len(steps)-1].Action != DeleteNamespace || steps[len(steps)-1].Ref.Namespace != "old" {
				t.Fatalf("plan = %#v, want create/replace prefix ending at old deletion", steps)
			}
			deleteCount := 0
			for _, step := range steps {
				if step.Action == DeleteNamespace {
					deleteCount++
				}
			}
			if deleteCount != 1 {
				t.Fatalf("plan = %#v, want exactly one observation frontier", steps)
			}
			buffer := steps[0]
			if buffer.Action != ModifyWorkspaceFixed || !buffer.Temporary || buffer.To.Limit-buffer.To.TotalFixed() != actual.Workspace.Limit-actual.Workspace.TotalFixed() {
				t.Fatalf("first step = %#v, want pure active-fixed topology buffer preserving elastic CU", buffer)
			}
		})
	}
}

func TestPlanAuthoritativeTopologyFallbackUsesPureBufferWithPartialFixedHeadroom(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(fmt.Sprintf("ha=%t", ha), func(t *testing.T) {
			actual, desired := authoritativePartialHeadroomRenameFixture(ha)
			endpoint := actual.Workspace.Limit

			steps, err := PlanAuthoritative(actual, desired)
			if err != nil {
				t.Fatal(err)
			}

			createIndex := -1
			var firstAboveEndpoint *Step
			for i := range steps {
				step := &steps[i]
				if step.Action == CreateNamespace && createIndex < 0 {
					createIndex = i
				}
				if isWorkspaceStep(*step) && step.To.Limit > endpoint && firstAboveEndpoint == nil {
					firstAboveEndpoint = step
				}
				if createIndex < 0 && (step.Action == EnableWorkspaceElastic || step.Action == ModifyWorkspaceElastic) {
					t.Fatalf("topology fallback changed elastic before Create: %#v", steps)
				}
			}
			if createIndex < 0 || firstAboveEndpoint == nil {
				t.Fatalf("plan = %#v, want endpoint-exceeding buffer followed by Create", steps)
			}
			buffer := *firstAboveEndpoint
			if buffer.Action != ModifyWorkspaceFixed || !buffer.Temporary {
				t.Fatalf("first endpoint-exceeding Workspace step = %#v, want temporary fixed topology buffer", buffer)
			}
			if buffer.To.TotalFixed()-buffer.From.TotalFixed() != minimumNamespaceCU || buffer.To.Limit-buffer.From.Limit != minimumNamespaceCU {
				t.Fatalf("buffer = %#v, want P+1 CU and L+1 CU", buffer)
			}
			if buffer.To.AsCapacity().Elastic() != buffer.From.AsCapacity().Elastic() {
				t.Fatalf("buffer = %#v, want elastic component unchanged", buffer)
			}
			if ha && (buffer.To.FixedCU != 0 || buffer.To.CrossZoneFixedCU-buffer.From.CrossZoneFixedCU != minimumNamespaceCU) {
				t.Fatalf("HA buffer changed the wrong fixed pool: %#v", buffer)
			}
			if !ha && (buffer.To.CrossZoneFixedCU != 0 || buffer.To.FixedCU-buffer.From.FixedCU != minimumNamespaceCU) {
				t.Fatalf("single-zone buffer changed the wrong fixed pool: %#v", buffer)
			}
		})
	}
}

func TestPlanAuthoritativeRetainedDesiredDoesNotProveRenameReplacement(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(fmt.Sprintf("ha=%t", ha), func(t *testing.T) {
			actual := authoritativeTree(8,
				authoritativeNamespace("keep-a", 2, 2, 0),
				authoritativeNamespace("keep-b", 2, 2, 0),
				namespaceWithQueues("old", 4, 4,
					Queue{Name: "default-queue", Capacity: capacity(4, 4), Used: 1.5},
					Queue{Name: "custom-busy", Capacity: capacity(2, 2), Used: 0.5},
				),
			)
			desired := authoritativeTree(8,
				authoritativeNamespace("keep-a", 2, 2, 0),
				authoritativeNamespace("keep-b", 2, 2, 0),
				authoritativeNamespace("new", 4, 4, 0),
			)
			actual.Workspace.Used = 2
			setAuthoritativeTreeHAForTest(&actual, ha)
			setAuthoritativeTreeHAForTest(&desired, ha)

			minimum, err := minimumAuthoritativePlanSteps(context.Background(), actual, desired)
			if err != nil {
				t.Fatal(err)
			}
			if minimum != 2 {
				t.Fatalf("minimum steps=%d, want conservative Create+Delete lower bound without replacement provenance", minimum)
			}

			steps, err := PlanAuthoritative(actual, desired)
			if err != nil {
				t.Fatal(err)
			}
			if len(steps) == 0 || steps[0].Action != ModifyWorkspaceFixed || !steps[0].Temporary {
				t.Fatalf("steps=%#v, want a pure topology buffer before the missing replacement, never source Delete", steps)
			}
			buffer := steps[0]
			if buffer.To.Limit-buffer.From.Limit != minimumNamespaceCU || buffer.To.TotalFixed()-buffer.From.TotalFixed() != minimumNamespaceCU || buffer.To.AsCapacity().Elastic() != buffer.From.AsCapacity().Elastic() {
				t.Fatalf("first step=%#v, want exactly one pure active-fixed CU of topology headroom", buffer)
			}
		})
	}
}

func TestPlanAuthoritativeNetGrowthRenameUsesAtMostOneCUWithoutGuessingReplacement(t *testing.T) {
	for _, sources := range []int{1, 2} {
		for _, ha := range []bool{false, true} {
			t.Run(fmt.Sprintf("sources=%d/ha=%t", sources, ha), func(t *testing.T) {
				actual, desired := authoritativeNetGrowthRenameFixture(sources, ha)
				endpoint := actual.Workspace.Limit

				steps, err := PlanAuthoritative(actual, desired)
				if err != nil {
					t.Fatal(err)
				}
				if len(steps) < 2 || steps[0].Action != CreateNamespace || steps[0].Ref.Namespace != "new-00" || steps[len(steps)-1].Action != DeleteNamespace || steps[len(steps)-1].Ref.Namespace != "old-00" {
					t.Fatalf("steps=%#v, want deterministic Create/buffer prefix ending at the smallest safe source Delete frontier", steps)
				}

				state := cloneTree(actual)
				usedBuffer := false
				for _, step := range steps {
					if isWorkspaceStep(step) {
						if step.To.Limit > endpoint+minimumNamespaceCU {
							t.Fatalf("replacement prefix exceeded operation-entry C0+1: endpoint=%v steps=%#v", endpoint.Float64(), steps)
						}
						if step.To.Limit > endpoint {
							usedBuffer = true
							if step.Action != ModifyWorkspaceFixed || !step.Temporary || step.To.AsCapacity().Elastic() != step.From.AsCapacity().Elastic() {
								t.Fatalf("endpoint-exceeding step is not a pure fixed topology buffer: %#v", step)
							}
						}
					}
					if step.Action == DeleteNamespace {
						missing, undeclared := authoritativeTopologyDebtForTest(state, desired)
						if missing > 0 && undeclared <= missing {
							t.Fatalf("Delete %#v guessed replacement provenance with missing=%d undeclared=%d; steps=%#v", step, missing, undeclared, steps)
						}
					}
					applyCandidate(&state, candidate{step: step})
				}
				if !usedBuffer {
					t.Fatalf("steps=%#v, want approved conservative C0+1 buffer for provenance-ambiguous net growth", steps)
				}
			})
		}
	}
}

func TestPlanAuthoritativeUnknownPeakReusesHeadroomOrFailsClosed(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(fmt.Sprintf("no-headroom/ha=%t", ha), func(t *testing.T) {
			actual, desired := authoritativeUnknownPeakFixture(ha, false)
			steps, err := PlanAuthoritative(actual, desired)
			if steps != nil || err == nil || !strings.Contains(err.Error(), "no safe capacity transition") {
				t.Fatalf("steps=%#v error=%v, want zero-step fail-closed at an unproven imported peak", steps, err)
			}
		})

		t.Run(fmt.Sprintf("visible-headroom/ha=%t", ha), func(t *testing.T) {
			actual, desired := authoritativeUnknownPeakFixture(ha, true)
			steps, err := PlanAuthoritative(actual, desired)
			if err != nil {
				t.Fatal(err)
			}
			if len(steps) == 0 || steps[0].Action != CreateNamespace || steps[0].Ref.Namespace != "new" {
				t.Fatalf("steps=%#v, want visible headroom reused for replacement Create", steps)
			}
			for _, step := range steps {
				if isWorkspaceStep(step) && step.To.Limit > actual.Workspace.Limit {
					t.Fatalf("unknown peak was stacked instead of reused: actual=%v steps=%#v", actual.Workspace.Limit.Float64(), steps)
				}
			}
		})
	}
}

func TestPlanAuthoritativeDeterministicRenameWorkScalesByFixtureSize(t *testing.T) {
	for _, pairs := range []int{2, 4, 6, 8} {
		for _, ha := range []bool{false, true} {
			t.Run(fmt.Sprintf("pairs=%d/ha=%t", pairs, ha), func(t *testing.T) {
				actual, desired := authoritativeRenameWorkFixture(pairs, ha)
				steps, stats, err := planAuthoritativeWithLimits(context.Background(), actual, desired, plannerLimits{})
				if err != nil {
					t.Fatal(err)
				}
				if stats.WorkUnits > 512*(pairs+1)*(pairs+1) {
					t.Fatalf("work units=%d, want polynomial bound for %d pairs", stats.WorkUnits, pairs)
				}
				reversedActual := cloneTree(actual)
				reversedDesired := cloneTree(desired)
				reverseNamespaces(reversedActual.Namespaces)
				reverseNamespaces(reversedDesired.Namespaces)
				reversed, reversedStats, err := planAuthoritativeWithLimits(context.Background(), reversedActual, reversedDesired, plannerLimits{})
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(steps, reversed) {
					t.Fatalf("plan depends on input order: %#v != %#v", steps, reversed)
				}
				if stats != reversedStats {
					t.Fatalf("planner stats depend on input order: %#v != %#v", stats, reversedStats)
				}
				if len(steps) > 8*pairs+8 {
					t.Fatalf("steps=%d, want linear mutation witness for %d pairs", len(steps), pairs)
				}
				if steps[len(steps)-1].Action != DeleteNamespace {
					t.Fatalf("plan does not stop at Delete frontier: %#v", steps)
				}
			})
		}
	}
}

func TestPlanAuthoritativeExhaustiveSmallCompositionStates(t *testing.T) {
	cases := 0
	for _, ha := range []bool{false, true} {
		for actualFixed := CU(2); actualFixed <= 8; actualFixed += 2 {
			for actualElastic := CU(0); actualElastic <= 6; actualElastic += 2 {
				for childFixed := CU(2); childFixed <= actualFixed; childFixed += 2 {
					for childLimit := childFixed; childLimit <= actualFixed+actualElastic; childLimit += 2 {
						actual := authoritativeCompositionTree(ha, actualFixed, actualFixed+actualElastic, childFixed)
						actual.Namespaces[0].Capacity.Limit = childLimit
						actual.Namespaces[0].Queues[0].Capacity.Limit = childLimit
						for desiredFixed := CU(2); desiredFixed <= 8; desiredFixed += 2 {
							for desiredElastic := CU(0); desiredElastic <= 6; desiredElastic += 2 {
								if actualElastic > 0 && desiredElastic == 0 {
									continue
								}
								desired := authoritativeCompositionTree(ha, desiredFixed, desiredFixed+desiredElastic, desiredFixed)
								steps, err := PlanAuthoritative(actual, desired)
								if err != nil {
									t.Fatalf("ha=%t actual=P%v/E%v child=F%v/L%v desired=P%v/E%v: %v", ha, actualFixed.Float64(), actualElastic.Float64(), childFixed.Float64(), childLimit.Float64(), desiredFixed.Float64(), desiredElastic.Float64(), err)
								}
								endpoint := maxCU(actual.Workspace.Limit, desired.Workspace.Limit)
								for _, step := range steps {
									if isWorkspaceStep(step) && step.To.Limit > endpoint+minimumNamespaceCU {
										t.Fatalf("ha=%t plan exceeded one-CU ceiling: %#v", ha, steps)
									}
								}
								if state := applyPlanForTest(t, actual, steps); !sameCapacityTree(state, desired) {
									t.Fatalf("ha=%t actual=%#v desired=%#v steps=%#v", ha, actual, desired, steps)
								}
								cases++
							}
						}
					}
				}
			}
		}
	}
	if cases < 1_000 {
		t.Fatalf("exercised only %d cases", cases)
	}
}

func TestPlanAuthoritativeContextAndBudgetFailClosed(t *testing.T) {
	actual := authoritativeCompositionTree(false, 8, 12, 8)
	desired := authoritativeCompositionTree(false, 4, 12, 4)

	t.Run("pre-canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := PlanAuthoritativeContext(ctx, actual, desired)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context canceled", err)
		}
	})

	t.Run("canceled during work", func(t *testing.T) {
		ctx := &cancelAfterChecksContext{Context: context.Background(), remaining: 5}
		_, err := PlanAuthoritativeContext(ctx, actual, desired)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want context canceled", err)
		}
	})

	t.Run("step budget", func(t *testing.T) {
		steps, _, err := planAuthoritativeWithLimits(context.Background(), actual, desired, plannerLimits{MaxSteps: 1, MaxWorkUnits: 1_000_000})
		var budgetErr *PlannerBudgetError
		if steps != nil || !errors.As(err, &budgetErr) || budgetErr.Kind != "steps" || budgetErr.Phase != "preflight" || budgetErr.Steps != 0 {
			t.Fatalf("error = %T %v, want preflight step PlannerBudgetError before scheduling", err, err)
		}
	})

	t.Run("work budget", func(t *testing.T) {
		steps, _, err := planAuthoritativeWithLimits(context.Background(), actual, desired, plannerLimits{MaxSteps: 4096, MaxWorkUnits: 7})
		var budgetErr *PlannerBudgetError
		if steps != nil || !errors.As(err, &budgetErr) || budgetErr.Kind != "work" || budgetErr.Phase != "endpoint" {
			t.Fatalf("error = %T %v, want work PlannerBudgetError", err, err)
		}
	})

	t.Run("preflight node budget", func(t *testing.T) {
		large := authoritativeTree(8,
			authoritativeNamespace("a", 2, 2, 0),
			authoritativeNamespace("b", 2, 2, 0),
			authoritativeNamespace("c", 2, 2, 0),
			authoritativeNamespace("d", 2, 2, 0),
		)
		_, _, err := planAuthoritativeWithLimits(context.Background(), large, cloneTree(large), plannerLimits{MaxSteps: 4096, MaxWorkUnits: 1})
		var budgetErr *PlannerBudgetError
		if !errors.As(err, &budgetErr) || budgetErr.Kind != "work" || budgetErr.Phase != "preflight" {
			t.Fatalf("error = %T %v, want preflight work PlannerBudgetError", err, err)
		}
	})
}

func TestScheduleAuthoritativePhaseCapsWitnessAllocationAtStepBudget(t *testing.T) {
	state := authoritativeCompositionTree(false, 2, 2, 2)
	budget, err := newPlannerBudget(context.Background(), state, state, plannerLimits{MaxSteps: 1, MaxWorkUnits: 1_000_000})
	if err != nil {
		t.Fatal(err)
	}
	// A synthetic but safe estimate reproduces the allocation bug without
	// asking the shared test process to reserve the schema-maximum witness.
	budget.estimatedBound = 1_024

	result, err := scheduleAuthoritativePhase(cloneTree(state), cloneTree(state), state.Workspace.Limit, state.Workspace.Limit, "endpoint", false, budget)
	if err != nil {
		t.Fatal(err)
	}
	if got := cap(result.steps); got > budget.limits.MaxSteps {
		t.Fatalf("witness capacity = %d, want hard cap <= MaxSteps %d", got, budget.limits.MaxSteps)
	}
}

func TestPlannerEstimatedBoundSaturatesInsteadOfOverflowing(t *testing.T) {
	actual := authoritativeCompositionTree(false, 2, 2, 2)
	desired := cloneTree(actual)
	actual.Workspace.FixedCU = CU(-1 << 63)
	desired.Workspace.FixedCU = CU(1<<63 - 1)

	budget, err := newPlannerBudget(context.Background(), actual, desired, plannerLimits{MaxSteps: 1, MaxWorkUnits: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := int(^uint(0) >> 1)
	if budget.estimatedBound != want {
		t.Fatalf("estimated bound = %d, want saturated max int %d", budget.estimatedBound, want)
	}
}

func TestPlanAuthoritativeRejectsProvableStepBudgetBeforeScheduling(t *testing.T) {
	// This is deliberately far below the public schema maximum. It safely
	// reproduces the missing preflight gate while still making the old witness
	// allocation visibly larger than the one-step hard budget.
	actual := authoritativeCompositionTree(false, 8_192, 8_194, 8_192)
	desired := authoritativeCompositionTree(false, 2, 8_194, 2)

	steps, stats, err := planAuthoritativeWithLimits(context.Background(), actual, desired, plannerLimits{MaxSteps: 1, MaxWorkUnits: 1_000_000})
	var budgetErr *PlannerBudgetError
	if steps != nil || !errors.As(err, &budgetErr) {
		t.Fatalf("steps=%#v error=%T %v, want nil steps and PlannerBudgetError", steps, err, err)
	}
	if budgetErr.Kind != "steps" || budgetErr.Phase != "preflight" || stats.Steps != 0 {
		t.Fatalf("budget error=%#v stats=%#v, want preflight step rejection before scheduling", budgetErr, stats)
	}
}

func TestPlanAuthoritativeSchemaMaximumFailsClosedWithinFixedStepBudget(t *testing.T) {
	actual := authoritativeCompositionTree(false, schemaMaximumCUHalfUnits, schemaMaximumCUHalfUnits, schemaMaximumCUHalfUnits)
	desired := authoritativeCompositionTree(false, minimumNamespaceCU, schemaMaximumCUHalfUnits, minimumNamespaceCU)

	steps, stats, err := planAuthoritativeWithLimits(context.Background(), actual, desired, plannerLimits{MaxSteps: 1, MaxWorkUnits: 1_000_000})
	var budgetErr *PlannerBudgetError
	if steps != nil || !errors.As(err, &budgetErr) {
		t.Fatalf("steps=%#v error=%T %v, want nil steps and PlannerBudgetError", steps, err, err)
	}
	if budgetErr.Kind != "steps" || budgetErr.Phase != "preflight" || stats.Steps != 0 {
		t.Fatalf("budget error=%#v stats=%#v, want schema-maximum preflight rejection before allocation", budgetErr, stats)
	}
	if budgetErr.EstimatedBound <= budgetErr.Steps {
		t.Fatalf("estimated bound = %d, want diagnostic estimate larger than consumed steps", budgetErr.EstimatedBound)
	}
}

func TestPlanAuthoritativeSchemaMaximumRenameRejectsUnserializableTopologyBuffer(t *testing.T) {
	for _, ha := range []bool{false, true} {
		t.Run(fmt.Sprintf("ha=%t", ha), func(t *testing.T) {
			actual := authoritativeCompositionTree(ha, schemaMaximumCUHalfUnits, schemaMaximumCUHalfUnits, schemaMaximumCUHalfUnits)
			actual.Namespaces[0].Name = "old"
			desired := cloneTree(actual)
			desired.Namespaces[0].Name = "new"

			steps, err := PlanAuthoritative(actual, desired)
			if steps != nil || err == nil || !strings.Contains(err.Error(), "no safe capacity transition") {
				t.Fatalf("steps=%#v error=%v, want nil witness and terminal no-safe error", steps, err)
			}
		})
	}
}

func TestPlanAuthoritativeAtomicChildActionsFitExactFourStepBudget(t *testing.T) {
	actual, desired := authoritativeAtomicDonorReceiverFixture()
	minimum, err := minimumAuthoritativePlanSteps(context.Background(), actual, desired)
	if err != nil {
		t.Fatal(err)
	}
	if minimum != 4 {
		t.Fatalf("minimum steps = %d, want four atomic namespace/queue API actions", minimum)
	}

	steps, _, err := planAuthoritativeWithLimits(context.Background(), actual, desired, plannerLimits{MaxSteps: 4, MaxWorkUnits: 1_000_000})
	if err != nil {
		t.Fatalf("four-step safe plan was rejected: %v", err)
	}
	if len(steps) != 4 {
		t.Fatalf("steps=%#v, want four atomic child actions", steps)
	}
	want := []struct {
		action    Action
		namespace string
		from      Allocation
		to        Allocation
	}{
		{ModifyQueue, "donor", Allocation{FixedCU: 8, Limit: 8}, Allocation{FixedCU: 4, Limit: 4}},
		{ModifyNamespace, "donor", Allocation{FixedCU: 8, Limit: 8}, Allocation{FixedCU: 4, Limit: 4}},
		{ModifyNamespace, "receiver", Allocation{FixedCU: 4, Limit: 4}, Allocation{FixedCU: 8, Limit: 8}},
		{ModifyQueue, "receiver", Allocation{FixedCU: 4, Limit: 4}, Allocation{FixedCU: 8, Limit: 8}},
	}
	for i, expected := range want {
		if steps[i].Action != expected.action || steps[i].Ref.Namespace != expected.namespace || steps[i].From != expected.from || steps[i].To != expected.to {
			t.Fatalf("step %d = %#v, want %s %s %#v -> %#v", i, steps[i], expected.action, expected.namespace, expected.from, expected.to)
		}
	}
	if state := applyPlanForTest(t, actual, steps); !sameCapacityTree(state, desired) {
		t.Fatalf("four-step witness did not converge safely: %#v", steps)
	}
}

func TestMinimumAuthoritativePlanStepsDoesNotOverestimateDeleteUsedOrHeadroomPaths(t *testing.T) {
	t.Run("busy delete frontier", func(t *testing.T) {
		actual := authoritativeTree(12,
			authoritativeNamespace("keep", 4, 4, 1),
			authoritativeNamespace("drop", 8, 8, 4),
		)
		actual.Workspace.Used = 5
		desired := authoritativeTree(2, authoritativeNamespace("keep", 2, 2, 0))

		minimum, err := minimumAuthoritativePlanSteps(context.Background(), actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if minimum != 1 {
			t.Fatalf("minimum steps = %d, want only the observable Delete frontier", minimum)
		}
		steps, _, err := planAuthoritativeWithLimits(context.Background(), actual, desired, plannerLimits{MaxSteps: 1, MaxWorkUnits: 1_000_000})
		if err != nil || len(steps) != 1 || steps[0].Action != DeleteNamespace {
			t.Fatalf("steps=%#v error=%v, want one busy Delete frontier within budget", steps, err)
		}
	})

	t.Run("rename without headroom", func(t *testing.T) {
		actual := authoritativeTree(4, authoritativeNamespace("old", 4, 4, 0))
		desired := authoritativeTree(4, authoritativeNamespace("new", 4, 4, 0))
		minimum, err := minimumAuthoritativePlanSteps(context.Background(), actual, desired)
		if err != nil {
			t.Fatal(err)
		}
		if minimum != 2 {
			t.Fatalf("minimum steps = %d, want conservative Create+Delete lower bound below the three-step buffer witness", minimum)
		}
	})
}

func TestPlannerBudgetCountersFailClosedOnIntegerOverflow(t *testing.T) {
	maximum := int(^uint(0) >> 1)
	budget := &plannerBudget{
		ctx:    context.Background(),
		limits: plannerLimits{MaxSteps: maximum, MaxWorkUnits: maximum},
		stats:  plannerStats{Steps: maximum, WorkUnits: maximum},
	}
	for name, call := range map[string]func() error{
		"steps": func() error { return budget.addStep("overflow") },
		"work":  func() error { return budget.addWork("overflow", 1) },
	} {
		t.Run(name, func(t *testing.T) {
			var budgetErr *PlannerBudgetError
			if err := call(); !errors.As(err, &budgetErr) || budgetErr.Kind != name {
				t.Fatalf("error=%T %v, want %s PlannerBudgetError", err, err, name)
			}
		})
	}
}

type cancelAfterChecksContext struct {
	context.Context
	remaining int
}

func (c *cancelAfterChecksContext) Err() error {
	c.remaining--
	if c.remaining <= 0 {
		return context.Canceled
	}
	return nil
}

func (c *cancelAfterChecksContext) Deadline() (time.Time, bool) { return time.Time{}, false }

func authoritativeRenameWorkFixture(pairs int, ha bool) (Tree, Tree) {
	actualNamespaces := make([]Namespace, 0, pairs*2+1)
	desiredNamespaces := make([]Namespace, 0, pairs*2+1)
	for i := 0; i < pairs; i++ {
		donorName := fmt.Sprintf("donor-%02d", i)
		receiverName := fmt.Sprintf("receiver-%02d", i)
		donorActual := authoritativeNamespace(donorName, 4, 4, 0)
		donorDesired := authoritativeNamespace(donorName, 2, 2, 0)
		receiverActual := authoritativeNamespace(receiverName, 2, 2, 0)
		receiverDesired := authoritativeNamespace(receiverName, 4, 4, 0)
		for _, namespace := range []*Namespace{&donorActual, &donorDesired, &receiverActual, &receiverDesired} {
			namespace.CrossZone = ha
		}
		actualNamespaces = append(actualNamespaces, donorActual, receiverActual)
		desiredNamespaces = append(desiredNamespaces, donorDesired, receiverDesired)
	}
	old := authoritativeNamespace("old", 2, 2, 0)
	newNamespace := authoritativeNamespace("new", 2, 2, 0)
	old.CrossZone = ha
	newNamespace.CrossZone = ha
	actualNamespaces = append(actualNamespaces, old)
	desiredNamespaces = append(desiredNamespaces, newNamespace)
	workspaceFixed := CU(pairs*6 + 2)
	actual := Tree{ChargeType: "PRE", Workspace: WorkspaceCapacity{HA: ha, Limit: workspaceFixed}, Namespaces: actualNamespaces}
	desired := Tree{ChargeType: "PRE", Workspace: WorkspaceCapacity{HA: ha, Limit: workspaceFixed}, Namespaces: desiredNamespaces}
	if ha {
		actual.Workspace.CrossZoneFixedCU = workspaceFixed
		desired.Workspace.CrossZoneFixedCU = workspaceFixed
	} else {
		actual.Workspace.FixedCU = workspaceFixed
		desired.Workspace.FixedCU = workspaceFixed
	}
	return actual, desired
}

func authoritativePartialHeadroomRenameFixture(ha bool) (Tree, Tree) {
	// External CU: actual P5/E0/L5 with old F4/L5 leaves one CU fixed
	// headroom and zero max headroom. Desired is P4/E1/L5 with new F4/L5.
	actual := authoritativeCompositionTree(ha, 10, 10, 8)
	actual.Namespaces[0].Name = "old"
	desired := authoritativeCompositionTree(ha, 8, 10, 8)
	desired.Namespaces[0].Name = "new"
	return actual, desired
}

func authoritativeAtomicDonorReceiverFixture() (Tree, Tree) {
	// External CU: P=L=6. The donor releases 2 CU before the receiver
	// consumes it. Namespace and queue APIs each carry fixed+limit atomically.
	actual := authoritativeTree(12,
		authoritativeNamespace("donor", 8, 8, 2),
		authoritativeNamespace("receiver", 4, 4, 0),
	)
	actual.Namespaces[0].Queues[0].Used = 2
	desired := authoritativeTree(12,
		authoritativeNamespace("donor", 4, 4, 0),
		authoritativeNamespace("receiver", 8, 8, 0),
	)
	return actual, desired
}

func authoritativeNetGrowthRenameFixture(sources int, ha bool) (Tree, Tree) {
	const desiredNamespaces = 3
	workspaceCapacity := CU(desiredNamespaces) * minimumNamespaceCU
	actual := Tree{
		ChargeType: "PRE",
		Workspace:  WorkspaceCapacity{HA: ha, Limit: workspaceCapacity},
		Namespaces: make([]Namespace, 0, sources),
	}
	desired := Tree{
		ChargeType: "PRE",
		Workspace:  WorkspaceCapacity{HA: ha, Limit: workspaceCapacity},
		Namespaces: make([]Namespace, 0, desiredNamespaces),
	}
	if ha {
		actual.Workspace.CrossZoneFixedCU = workspaceCapacity
		desired.Workspace.CrossZoneFixedCU = workspaceCapacity
	} else {
		actual.Workspace.FixedCU = workspaceCapacity
		desired.Workspace.FixedCU = workspaceCapacity
	}
	for i := 0; i < sources; i++ {
		namespace := authoritativeNamespace(fmt.Sprintf("old-%02d", i), minimumNamespaceCU, minimumNamespaceCU, 0)
		namespace.CrossZone = ha
		actual.Namespaces = append(actual.Namespaces, namespace)
	}
	for i := 0; i < desiredNamespaces; i++ {
		namespace := authoritativeNamespace(fmt.Sprintf("new-%02d", i), minimumNamespaceCU, minimumNamespaceCU, 0)
		namespace.CrossZone = ha
		desired.Namespaces = append(desired.Namespaces, namespace)
	}
	return actual, desired
}

func authoritativeUnknownPeakFixture(ha, withHeadroom bool) (Tree, Tree) {
	oldCapacity := CU(6)
	if withHeadroom {
		oldCapacity = 4
	}
	actual := authoritativeTree(8,
		authoritativeNamespace("keep", 2, 2, 0),
		authoritativeNamespace("old", oldCapacity, oldCapacity, 0),
	)
	desired := authoritativeTree(6,
		authoritativeNamespace("keep", 2, 2, 0),
		authoritativeNamespace("new", 4, 4, 0),
	)
	setAuthoritativeTreeHAForTest(&actual, ha)
	setAuthoritativeTreeHAForTest(&desired, ha)
	return actual, desired
}

func setAuthoritativeTreeHAForTest(tree *Tree, ha bool) {
	tree.Workspace.HA = ha
	if ha {
		tree.Workspace.CrossZoneFixedCU = tree.Workspace.FixedCU
		tree.Workspace.FixedCU = 0
	}
	for i := range tree.Namespaces {
		tree.Namespaces[i].CrossZone = ha
	}
}

func authoritativeTopologyDebtForTest(actual, desired Tree) (missing, undeclared int) {
	for _, wanted := range desired.Namespaces {
		if namespaceByName(actual, wanted.Name) == nil {
			missing++
		}
	}
	for _, observed := range actual.Namespaces {
		if namespaceByName(desired, observed.Name) == nil {
			undeclared++
		}
	}
	return missing, undeclared
}

func reverseNamespaces(namespaces []Namespace) {
	for left, right := 0, len(namespaces)-1; left < right; left, right = left+1, right-1 {
		namespaces[left], namespaces[right] = namespaces[right], namespaces[left]
	}
}

func isWorkspaceStep(step Step) bool {
	switch step.Action {
	case ModifyWorkspaceFixed, ModifyWorkspacePostpaid, EnableWorkspaceElastic, ModifyWorkspaceElastic:
		return true
	default:
		return false
	}
}
