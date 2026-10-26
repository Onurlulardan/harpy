package planner

import (
	"errors"
	"math"
	"time"
)

const (
	HardRecordStatementCost     = 50_000.0
	HardRecordRequestCost       = 100_000.0
	HardExactCountStatementCost = 50_000.0
	HardAggregateStatementCost  = 100_000.0
	HardInteractiveTimeout      = 5 * time.Second
	HardExactCountTimeout       = 2 * time.Second
)

var ErrPolicyInvalid = errors.New("REST Query planner policy is invalid")

// Limits is copied into Policy; callers cannot mutate an established policy.
type Limits struct {
	RecordStatementCost     float64
	RecordRequestCost       float64
	ExactCountStatementCost float64
	AggregateStatementCost  float64
	InteractiveTimeout      time.Duration
	ExactCountTimeout       time.Duration
}

type Policy struct {
	limits Limits
}

func HardLimits() Limits {
	return Limits{
		RecordStatementCost:     HardRecordStatementCost,
		RecordRequestCost:       HardRecordRequestCost,
		ExactCountStatementCost: HardExactCountStatementCost,
		AggregateStatementCost:  HardAggregateStatementCost,
		InteractiveTimeout:      HardInteractiveTimeout,
		ExactCountTimeout:       HardExactCountTimeout,
	}
}

func NewPolicy(limits Limits) (Policy, error) {
	hard := HardLimits()
	if !validCostLimit(limits.RecordStatementCost, hard.RecordStatementCost) ||
		!validCostLimit(limits.RecordRequestCost, hard.RecordRequestCost) ||
		!validCostLimit(limits.ExactCountStatementCost, hard.ExactCountStatementCost) ||
		!validCostLimit(limits.AggregateStatementCost, hard.AggregateStatementCost) ||
		!validTimeout(limits.InteractiveTimeout, hard.InteractiveTimeout) ||
		!validTimeout(limits.ExactCountTimeout, hard.ExactCountTimeout) {
		return Policy{}, ErrPolicyInvalid
	}
	return Policy{limits: limits}, nil
}

func HardPolicy() Policy { return Policy{limits: HardLimits()} }

func (policy Policy) Limits() Limits { return policy.limits }

func (policy Policy) valid() bool {
	_, err := NewPolicy(policy.limits)
	return err == nil
}

func validCostLimit(value, maximum float64) bool {
	return value > 0 && value <= maximum && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validTimeout(value, maximum time.Duration) bool {
	return value > 0 && value <= maximum
}
