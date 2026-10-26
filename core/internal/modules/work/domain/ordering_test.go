package domain

import "testing"

func TestPlanPositionUsesMidpointAndBoundedRebalance(t *testing.T) {
	first := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9870")
	second := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9871")
	moving := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9872")
	plan, err := PlanPosition([]RankedTask{{first, 10}, {second, 20}}, moving,
		MoveNeighbors{Before: &first, After: &second})
	if err != nil || plan.Position != 15 || len(plan.Rebalanced) != 0 {
		t.Fatalf("midpoint plan = %+v, %v", plan, err)
	}
	plan, err = PlanPosition([]RankedTask{{first, 10}, {second, 11}}, moving,
		MoveNeighbors{Before: &first, After: &second})
	if err != nil || plan.Position != 1536 || len(plan.Rebalanced) != 2 {
		t.Fatalf("rebalance plan = %+v, %v", plan, err)
	}
}

func TestPlanPositionRejectsNonAdjacentNeighbors(t *testing.T) {
	first := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9870")
	middle := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9871")
	last := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9872")
	moving := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9873")
	_, err := PlanPosition([]RankedTask{{first, 1}, {middle, 2}, {last, 3}}, moving,
		MoveNeighbors{Before: &first, After: &last})
	if err == nil {
		t.Fatal("non-adjacent neighbors were accepted")
	}
}

func mustID(t *testing.T, value string) ID {
	t.Helper()
	result, err := ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
