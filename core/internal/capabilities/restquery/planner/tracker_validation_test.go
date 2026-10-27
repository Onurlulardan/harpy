package planner

import (
	"errors"
	"math"
	"testing"
)

func TestTrackerRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	if _, err := NewTracker(Policy{}); !errors.Is(err, ErrPolicyInvalid) {
		t.Fatalf("zero policy error = %v", err)
	}
	tracker, _ := NewTracker(HardPolicy())
	if _, err := tracker.Admit(0, Cost{}); !errors.Is(err, ErrStatementClassInvalid) {
		t.Fatalf("invalid class error = %v", err)
	}
	invalidCost := Cost{value: math.NaN()}
	if _, err := tracker.Admit(StatementClassRecord, invalidCost); !errors.Is(err, ErrCostInvalid) {
		t.Fatalf("invalid cost error = %v", err)
	}
	if _, err := NewCost(math.Inf(1)); !errors.Is(err, ErrCostInvalid) {
		t.Fatalf("infinite cost error = %v", err)
	}
	if _, err := NewCost(-1); !errors.Is(err, ErrCostInvalid) {
		t.Fatalf("negative cost error = %v", err)
	}
	if _, err := NewCost(math.Copysign(0, -1)); !errors.Is(err, ErrCostInvalid) {
		t.Fatalf("negative zero cost error = %v", err)
	}
}

func TestLoweredPolicyAppliesToStatementAndRequest(t *testing.T) {
	t.Parallel()
	limits := HardLimits()
	limits.RecordStatementCost = 20
	limits.RecordRequestCost = 30
	policy, _ := NewPolicy(limits)
	tracker, _ := NewTracker(policy)
	tracker, err := tracker.Admit(StatementClassRecord, mustCost(t, 15))
	if err != nil {
		t.Fatal(err)
	}
	_, err = tracker.Admit(StatementClassInclude, mustCost(t, 16))
	requireQueryError(t, err, "QUERY_BUDGET_EXCEEDED", true)
	_, err = tracker.Admit(StatementClassInclude, mustCost(t, 21))
	requireQueryError(t, err, "QUERY_BUDGET_EXCEEDED", true)
}
