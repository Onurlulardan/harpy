package v1query

import (
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func buildOperationFilter(
	field, value, path string, operation resolvedOperation,
	collector *issueCollector,
) (api.RequestFilter, bool) {
	switch operation.kind {
	case legacyNotContains:
		child, valid := newPredicate(
			field, api.OperatorContains, []string{value}, path, collector,
		)
		if !valid {
			return api.RequestFilter{}, false
		}
		result, err := api.NewRequestNot(child)
		return checkedFilter(result, err, path, collector)
	case legacyIsEmpty:
		return buildEmptyFilter(field, true, path, collector)
	case legacyIsNotEmpty:
		return buildEmptyFilter(field, false, path, collector)
	case legacyDirect:
		values, valid := operationValues(operation.operator, value, path, collector)
		if !valid {
			return api.RequestFilter{}, false
		}
		return newPredicate(field, operation.operator, values, path, collector)
	default:
		collector.add(path, restqueryv1.IssueOperatorUnknown)
		return api.RequestFilter{}, false
	}
}

func buildEmptyFilter(
	field string, empty bool, path string, collector *issueCollector,
) (api.RequestFilter, bool) {
	firstOperator, secondOperator := api.OperatorIsNull, api.OperatorEq
	constructor := api.NewRequestOr
	if !empty {
		firstOperator, secondOperator = api.OperatorIsNotNull, api.OperatorNe
		constructor = api.NewRequestAnd
	}
	first, valid := newPredicate(field, firstOperator, nil, path, collector)
	if !valid {
		return api.RequestFilter{}, false
	}
	second, valid := newPredicate(field, secondOperator, []string{""}, path, collector)
	if !valid {
		return api.RequestFilter{}, false
	}
	result, err := constructor([]api.RequestFilter{first, second})
	return checkedFilter(result, err, path, collector)
}

func operationValues(
	operator api.Operator, raw, path string, collector *issueCollector,
) ([]string, bool) {
	arity, valid := operator.Arity()
	if !valid {
		collector.add(path, restqueryv1.IssueOperatorUnknown)
		return nil, false
	}
	switch arity {
	case api.PredicateArityZero:
		return nil, true
	case api.PredicateArityOne:
		return []string{raw}, true
	case api.PredicateArityList:
		values := splitCommaValues(raw)
		if operator == api.OperatorContainsAny {
			values = nonEmptyCommaValues(raw)
			if len(values) == 0 {
				return nil, false
			}
		}
		if len(values) > 100 {
			collector.addBudget(path, restqueryv1.IssueOperandCountExceeded)
			return nil, false
		}
		return values, true
	case api.PredicateArityPair:
		values := splitCommaValues(raw)
		if len(values) != 2 {
			collector.add(path, restqueryv1.IssueArityMismatch)
			return nil, false
		}
		return values, true
	default:
		collector.add(path, restqueryv1.IssueArityMismatch)
		return nil, false
	}
}

func splitCommaValues(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func newPredicate(
	field string, operator api.Operator, values []string, path string,
	collector *issueCollector,
) (api.RequestFilter, bool) {
	literals := make([]api.Literal, 0, len(values))
	for _, value := range values {
		literal, err := api.NewTextLiteral(value)
		if err != nil {
			collector.add(path, restqueryv1.IssueValueNotAllowed)
			return api.RequestFilter{}, false
		}
		literals = append(literals, literal)
	}
	result, err := api.NewRequestPredicate(field, operator, literals)
	return checkedFilter(result, err, path, collector)
}

func checkedFilter(
	filter api.RequestFilter, err error, path string, collector *issueCollector,
) (api.RequestFilter, bool) {
	if err != nil {
		collector.addBudget(path, restqueryv1.IssueASTNodeCountExceeded)
		return api.RequestFilter{}, false
	}
	return filter, true
}
