package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

func (statement RecordStatement) ValidatePlan(plan resolved.AuthorizedPlan) error {
	binding, count, err := recordPlanBinding(plan)
	if statement.Validate() != nil || err != nil || binding != statement.binding ||
		count != statement.countMode {
		return ErrContract
	}
	return nil
}
