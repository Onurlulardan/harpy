package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func rootConditions(
	query resolved.RecordQuery,
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

func visibilityCondition(
	visibility resolved.ResourceVisibility, binder *binder,
) (string, error) {
	mode, err := visibility.Mode()
	if err != nil {
		return "", ErrContract
	}
	switch mode {
	case resolved.ResourceVisibilityAll:
		return "", nil
	case resolved.ResourceVisibilityNone:
		return "FALSE", nil
	case resolved.ResourceVisibilityPredicate:
		filter, present, err := visibility.Filter()
		if err != nil || !present {
			return "", ErrContract
		}
		return compileMandatoryFilter(filter, binder)
	default:
		return "", ErrContract
	}
}

func parenthesize(value string) string {
	return "(" + strings.TrimSpace(value) + ")"
}
