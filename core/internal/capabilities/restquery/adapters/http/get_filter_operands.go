package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func requiredGETFilterField(
	item *getFilterDocument, path string, collector *issueCollector,
) (string, bool) {
	if item.field == nil {
		collector.add(path+"/field", restqueryv1.IssueMemberRequired)
		return "", false
	}
	return decodeGETFieldName(*item.field, collector)
}

func requiredGETOperator(
	item *getFilterDocument, path string, collector *issueCollector,
) (api.Operator, bool) {
	if item.operator == nil {
		collector.add(path+"/op", restqueryv1.IssueMemberRequired)
		return 0, false
	}
	operator, found := api.ParseOperator(item.operator.value)
	if !found {
		collector.add(item.operator.pointer, restqueryv1.IssueOperatorUnknown)
		return 0, false
	}
	return operator, true
}

func decodeGETOperands(
	item *getFilterDocument,
	path string,
	operator api.Operator,
	collector *issueCollector,
) ([]api.Literal, bool) {
	arity, _ := operator.Arity()
	switch arity {
	case api.PredicateArityZero:
		return nil, rejectGETOperands(item, path, collector)
	case api.PredicateArityOne:
		valid := rejectGETValues(item, path, collector)
		if item.value == nil {
			collector.add(path+"/value", restqueryv1.IssueMemberRequired)
			return nil, false
		}
		literal, literalValid := decodeGETLiteral(*item.value, collector)
		return []api.Literal{literal}, valid && literalValid
	case api.PredicateArityList, api.PredicateArityPair:
		valid := rejectGETValue(item, path, collector)
		values, valuesValid := decodeGETValues(item, path, arity, collector)
		return values, valid && valuesValid
	default:
		return nil, false
	}
}

func decodeGETValues(
	item *getFilterDocument,
	path string,
	arity api.PredicateArity,
	collector *issueCollector,
) ([]api.Literal, bool) {
	valuesPath := path + "/values"
	if len(item.values) == 0 {
		collector.add(valuesPath, restqueryv1.IssueMemberRequired)
		return nil, false
	}
	indices, valid := orderedGETIndices(item.values, valuesPath, collector)
	if arity == api.PredicateArityList && len(indices) > 100 {
		collector.addBudget(valuesPath, restqueryv1.IssueOperandCountExceeded)
		valid = false
	}
	if arity == api.PredicateArityPair && len(indices) != 2 {
		collector.add(valuesPath, restqueryv1.IssueArityMismatch)
		valid = false
	}
	result := make([]api.Literal, 0, len(indices))
	for _, index := range indices {
		literal, literalValid := decodeGETLiteral(item.values[index], collector)
		if !literalValid {
			valid = false
			continue
		}
		result = append(result, literal)
	}
	return result, valid
}

func rejectGETOperands(
	item *getFilterDocument, path string, collector *issueCollector,
) bool {
	valueValid := rejectGETValue(item, path, collector)
	valuesValid := rejectGETValues(item, path, collector)
	return valueValid && valuesValid
}

func rejectGETValue(
	item *getFilterDocument, path string, collector *issueCollector,
) bool {
	if item.value != nil {
		collector.add(path+"/value", restqueryv1.IssueArityMismatch)
		return false
	}
	return true
}

func rejectGETValues(
	item *getFilterDocument, path string, collector *issueCollector,
) bool {
	if len(item.values) != 0 {
		collector.add(path+"/values", restqueryv1.IssueArityMismatch)
		return false
	}
	return true
}
