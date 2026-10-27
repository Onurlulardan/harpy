package planner

import "math"

// Tracker is an immutable per-request planner budget accumulator.
type Tracker struct {
	policy            Policy
	recordRequestCost float64
}

func NewTracker(policy Policy) (Tracker, error) {
	if !policy.valid() {
		return Tracker{}, ErrPolicyInvalid
	}
	return Tracker{policy: policy}, nil
}

func (tracker Tracker) Admit(class StatementClass, cost Cost) (Tracker, error) {
	statement, err := tracker.policy.Statement(class)
	if err != nil {
		return tracker, err
	}
	if !cost.valid() {
		return tracker, ErrCostInvalid
	}
	if cost.value > statement.maximumCost {
		return tracker, budgetError(class)
	}
	if !statement.sharesRequestBudget {
		return tracker, nil
	}
	next := tracker.recordRequestCost + cost.value
	if math.IsInf(next, 0) || next > tracker.policy.limits.RecordRequestCost {
		return tracker, budgetError(class)
	}
	result := tracker
	result.recordRequestCost = next
	return result, nil
}

func (tracker Tracker) RecordRequestCost() Cost {
	return Cost{value: tracker.recordRequestCost}
}
