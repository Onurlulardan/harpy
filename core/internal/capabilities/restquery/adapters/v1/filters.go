package v1query

import (
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeFilters(
	document v1Document, policy manifest, collector *issueCollector,
) *api.RequestFilter {
	joinWithOr := false
	if document.filterOperator != nil && document.filterOperator.value != "" {
		switch strings.ToUpper(document.filterOperator.value) {
		case "AND":
		case "OR":
			joinWithOr = true
		default:
			collector.add(
				document.filterOperator.path, restqueryv1.IssueOperatorUnknown,
			)
		}
	}
	children := make([]api.RequestFilter, 0, len(document.filters))
	seenFields := make(map[string]struct{}, len(document.filters))
	for _, raw := range document.orderedFilters() {
		binding, found := policy.filterFields[raw.field]
		if !found {
			collector.add(raw.path, restqueryv1.IssueFieldUnknown)
			continue
		}
		document.aliases.add(
			binding.aliasID, raw.path, restqueryv1.IssueFieldUnknown,
		)
		if _, duplicate := seenFields[binding.name]; duplicate {
			collector.add(raw.path, restqueryv1.IssueFieldDuplicate)
			continue
		}
		seenFields[binding.name] = struct{}{}
		filter, present := decodeOneFilter(
			raw, binding, policy, &document.aliases, collector,
		)
		if present {
			children = append(children, filter)
		}
	}
	if len(children) == 0 {
		return nil
	}
	if len(children) == 1 {
		return &children[0]
	}
	if addFilterBudgetIssues(children, collector) {
		return nil
	}
	var (
		result api.RequestFilter
		err    error
	)
	if joinWithOr {
		result, err = api.NewRequestOr(children)
	} else {
		result, err = api.NewRequestAnd(children)
	}
	if err != nil {
		collector.addBudget("/filters", restqueryv1.IssueASTNodeCountExceeded)
		return nil
	}
	return &result
}

func decodeOneFilter(
	raw *rawFilter, binding fieldBinding, policy manifest, aliases *aliasSet,
	collector *issueCollector,
) (api.RequestFilter, bool) {
	if raw.scalar != nil {
		if raw.scalar.value == "" {
			return api.RequestFilter{}, false
		}
		operation, _ := legacyOperator("equals")
		if !operationMatchesShape(operation, binding.shape) {
			collector.add(raw.scalar.path, restqueryv1.IssueOperatorNotAllowed)
			return api.RequestFilter{}, false
		}
		return buildOperationFilter(
			binding.name, raw.scalar.value, raw.scalar.path, operation, collector,
		)
	}
	if len(raw.array) != 0 {
		return decodeArrayFilter(raw, binding, collector)
	}
	if raw.operator == nil || raw.operator.value == "" {
		return api.RequestFilter{}, false
	}
	operation, found := resolveOperator(
		raw.operator.value, raw.operator.path, policy, aliases,
	)
	if !found {
		collector.add(raw.operator.path, restqueryv1.IssueOperatorUnknown)
		return api.RequestFilter{}, false
	}
	if !operationMatchesShape(operation, binding.shape) {
		collector.add(raw.operator.path, restqueryv1.IssueOperatorNotAllowed)
		return api.RequestFilter{}, false
	}
	if raw.value == nil || raw.value.value == "" {
		return api.RequestFilter{}, false
	}
	return buildOperationFilter(
		binding.name, raw.value.value, raw.value.path, operation, collector,
	)
}

func decodeArrayFilter(
	raw *rawFilter, binding fieldBinding, collector *issueCollector,
) (api.RequestFilter, bool) {
	path := raw.path
	if len(raw.array) != 0 {
		path = raw.array[0].path
	}
	if binding.shape != FilterShapeCollection {
		collector.add(path, restqueryv1.IssueOperatorNotAllowed)
		return api.RequestFilter{}, false
	}
	values := make([]string, 0, len(raw.array))
	for _, member := range raw.array {
		if member.value != "" {
			values = append(values, member.value)
		}
	}
	values = nonEmptyCommaValues(strings.Join(values, ","))
	if len(values) == 0 {
		return api.RequestFilter{}, false
	}
	if len(values) > 100 {
		collector.addBudget(path, restqueryv1.IssueOperandCountExceeded)
		return api.RequestFilter{}, false
	}
	return newPredicate(
		binding.name, api.OperatorContainsAny, values, path, collector,
	)
}

func nonEmptyCommaValues(raw string) []string {
	parts := splitCommaValues(raw)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
