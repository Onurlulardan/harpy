package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileFilter(filter resolved.Filter, binder *binder) (string, error) {
	kind, err := filter.Kind()
	if err != nil {
		return "", ErrContract
	}
	if kind == api.RequestFilterPredicate {
		predicate, present, predicateErr := filter.Predicate()
		if predicateErr != nil || !present {
			return "", ErrContract
		}
		field, fieldErr := predicate.Field()
		operator, operatorErr := predicate.Operator()
		values, valuesErr := predicate.Values()
		if fieldErr != nil || operatorErr != nil || valuesErr != nil {
			return "", ErrContract
		}
		return compilePredicate(field, operator, values, binder)
	}
	children, err := filter.Children()
	if err != nil {
		return "", ErrContract
	}
	compiled := make([]string, len(children))
	for index, child := range children {
		value, childErr := compileFilter(child, binder)
		if childErr != nil || strings.TrimSpace(value) == "" {
			return "", ErrContract
		}
		compiled[index] = parenthesize(value)
	}
	return compileBoolean(kind, compiled)
}

func compileBoolean(kind api.RequestFilterKind, children []string) (string, error) {
	switch kind {
	case api.RequestFilterAnd:
		if len(children) < 2 {
			return "", ErrContract
		}
		return strings.Join(children, " AND "), nil
	case api.RequestFilterOr:
		if len(children) < 2 {
			return "", ErrContract
		}
		return strings.Join(children, " OR "), nil
	case api.RequestFilterNot:
		if len(children) != 1 {
			return "", ErrContract
		}
		return "NOT " + children[0], nil
	default:
		return "", ErrContract
	}
}
