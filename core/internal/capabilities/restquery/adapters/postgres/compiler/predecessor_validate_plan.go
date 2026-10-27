package compiler

import (
	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func (statement PredecessorStatement) ValidatePlan(
	plan resolved.AuthorizedPlan, ordinal v1query.PredecessorOrdinal,
) error {
	base, predecessorIndex, err := v1PlanOrdinal(plan, ordinal)
	if statement.Validate() != nil || err != nil ||
		statement.predecessorIndex != predecessorIndex {
		return ErrContract
	}
	expected, err := v1PredecessorBinding(base, predecessorIndex)
	if err != nil || expected != statement.binding {
		return ErrContract
	}
	return nil
}

func v1PlanOrdinal(
	plan resolved.AuthorizedPlan, ordinal v1query.PredecessorOrdinal,
) (statementBinding, int64, error) {
	binding, err := v1BaseBinding(plan)
	index, indexErr := ordinal.Index()
	limit, limitErr := ordinal.Limit()
	if err != nil || indexErr != nil || limitErr != nil {
		return statementBinding{}, 0, ErrContract
	}
	spec, err := plan.Spec()
	if err != nil {
		return statementBinding{}, 0, ErrContract
	}
	query, found, err := spec.Record()
	if err != nil || !found {
		return statementBinding{}, 0, ErrContract
	}
	page, err := query.Page()
	if err != nil {
		return statementBinding{}, 0, ErrContract
	}
	size, err := page.Size()
	if err != nil || size != limit {
		return statementBinding{}, 0, ErrContract
	}
	return binding, index, nil
}

func v1BaseBinding(plan resolved.AuthorizedPlan) (statementBinding, error) {
	binding, count, err := recordPlanBinding(plan)
	if err != nil || count != api.CountModeExact {
		return statementBinding{}, ErrContract
	}
	spec, err := plan.Spec()
	if err != nil {
		return statementBinding{}, ErrContract
	}
	query, found, err := spec.Record()
	if err != nil || !found || validateRecordHasNoIncludes(query) != nil {
		return statementBinding{}, ErrContract
	}
	page, err := query.Page()
	if err != nil {
		return statementBinding{}, ErrContract
	}
	_, hasAfter, err := page.After()
	if err != nil || hasAfter {
		return statementBinding{}, ErrContract
	}
	return binding, nil
}
