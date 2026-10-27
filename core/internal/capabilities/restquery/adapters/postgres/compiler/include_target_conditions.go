package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func includeTargetConditions(
	plan resolved.AuthorizedPlan,
	input includeNodeInput,
	candidatePlaceholder string,
	binder *binder,
) ([]string, error) {
	stable, err := input.include.StableID()
	if err != nil || candidatePlaceholder == "" || binder == nil {
		return nil, ErrContract
	}
	expression, err := baseExpression(stable)
	if err != nil {
		return nil, ErrContract
	}
	conditions := []string{
		expression + " = ANY(CAST(" + candidatePlaceholder +
			" AS pg_catalog.uuid[]))",
	}
	target, targetErr := input.include.Target()
	path, pathErr := input.include.Path()
	pathValue, pathValueErr := path.Value()
	if targetErr != nil || pathErr != nil || pathValueErr != nil {
		return nil, ErrContract
	}
	visibility, err := rootVisibility(plan, target, pathValue)
	if err != nil {
		return nil, ErrContract
	}
	mandatory, err := visibilityCondition(visibility, binder)
	if err != nil {
		return nil, err
	}
	if mandatory != "" {
		conditions = append(conditions, parenthesize(mandatory))
	}
	where, present, err := input.include.Where()
	if err != nil {
		return nil, ErrContract
	}
	if present {
		compiled, err := compileFilter(where, binder)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, parenthesize(compiled))
	}
	return conditions, nil
}
