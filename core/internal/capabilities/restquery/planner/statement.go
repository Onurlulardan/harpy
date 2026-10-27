package planner

import "time"

type Statement struct {
	class               StatementClass
	maximumCost         float64
	timeout             time.Duration
	sharesRequestBudget bool
}

func (policy Policy) Statement(class StatementClass) (Statement, error) {
	if !policy.valid() {
		return Statement{}, ErrPolicyInvalid
	}
	if !class.Valid() {
		return Statement{}, ErrStatementClassInvalid
	}
	statement := Statement{class: class}
	switch class {
	case StatementClassRecord, StatementClassInclude, StatementClassV1PredecessorSeek:
		statement.maximumCost = policy.limits.RecordStatementCost
		statement.timeout = policy.limits.InteractiveTimeout
		statement.sharesRequestBudget = true
	case StatementClassExactCount:
		statement.maximumCost = policy.limits.ExactCountStatementCost
		statement.timeout = policy.limits.ExactCountTimeout
	case StatementClassAggregate:
		statement.maximumCost = policy.limits.AggregateStatementCost
		statement.timeout = policy.limits.InteractiveTimeout
	}
	return statement, nil
}

func (statement Statement) Class() StatementClass  { return statement.class }
func (statement Statement) MaximumCost() float64   { return statement.maximumCost }
func (statement Statement) Timeout() time.Duration { return statement.timeout }

func (statement Statement) SharesRequestBudget() bool {
	return statement.sharesRequestBudget
}

func (statement Statement) TimeoutError() error {
	if statement.class == StatementClassExactCount {
		return newQueryError(queryErrorCountTimeout, false)
	}
	if !statement.class.Valid() {
		return ErrStatementClassInvalid
	}
	return newQueryError(queryErrorTimeout, false)
}
