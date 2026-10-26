package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func aggregateConditions(
	query resolved.AggregateQuery,
	visibility resolved.ResourceVisibility,
	binder *binder,
) ([]string, error) {
	conditions := make([]string, 0, 3)
	mandatory, err := visibilityCondition(visibility, binder)
	if err != nil {
		return nil, err
	}
	if mandatory != "" {
		conditions = append(conditions, parenthesize(mandatory))
	}
	where, present, err := query.Where()
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
	search, err := query.Search()
	if err != nil {
		return nil, ErrContract
	}
	if len(search) != 0 {
		compiled, err := compileSearch(search, binder)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, parenthesize(compiled))
	}
	return conditions, nil
}
