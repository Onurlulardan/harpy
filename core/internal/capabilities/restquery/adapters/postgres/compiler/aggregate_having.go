package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileAggregateHaving(
	having resolved.Having, input aggregateInput, binder *binder,
) (string, error) {
	kind, err := having.Kind()
	if err != nil || binder == nil {
		return "", ErrContract
	}
	if kind == api.RequestHavingKindPredicate {
		predicate, present, err := having.Predicate()
		if err != nil || !present {
			return "", ErrContract
		}
		return compileAggregateHavingPredicate(predicate, input, binder)
	}
	children, err := having.Children()
	if err != nil {
		return "", ErrContract
	}
	compiled := make([]string, len(children))
	for index, child := range children {
		value, err := compileAggregateHaving(child, input, binder)
		if err != nil || strings.TrimSpace(value) == "" {
			return "", ErrContract
		}
		compiled[index] = parenthesize(value)
	}
	switch kind {
	case api.RequestHavingKindAnd:
		if len(compiled) < 2 {
			return "", ErrContract
		}
		return strings.Join(compiled, " AND "), nil
	case api.RequestHavingKindOr:
		if len(compiled) < 2 {
			return "", ErrContract
		}
		return strings.Join(compiled, " OR "), nil
	case api.RequestHavingKindNot:
		if len(compiled) != 1 {
			return "", ErrContract
		}
		return "NOT " + compiled[0], nil
	default:
		return "", ErrContract
	}
}
