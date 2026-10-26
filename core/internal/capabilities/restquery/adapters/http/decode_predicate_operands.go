package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodePredicateOperands(
	object rawObject,
	path string,
	operator api.Operator,
	operatorValid bool,
	collector *issueCollector,
) ([]api.Literal, bool) {
	value, hasValue := object["value"]
	values, hasValues := object["values"]
	if !operatorValid {
		return nil, false
	}
	arity, _ := operator.Arity()
	switch arity {
	case api.PredicateArityZero:
		return nil, rejectOperands(path, hasValue, hasValues, collector)
	case api.PredicateArityOne:
		valid := rejectValuesMember(path, hasValues, collector)
		if !hasValue {
			collector.add(rawMemberPath(path, "value"), restqueryv1.IssueMemberRequired)
			return nil, false
		}
		literal, literalValid := decodeLiteral(value, rawMemberPath(path, "value"), collector)
		return []api.Literal{literal}, valid && literalValid
	case api.PredicateArityList, api.PredicateArityPair:
		valid := rejectValueMember(path, hasValue, collector)
		if !hasValues {
			collector.add(rawMemberPath(path, "values"), restqueryv1.IssueMemberRequired)
			return nil, false
		}
		result, valuesValid := decodeLiteralArray(values, rawMemberPath(path, "values"), collector)
		valid = validateOperandCount(arity, len(result), path, collector) && valid
		return result, valuesValid && valid
	default:
		return nil, false
	}
}

func rejectOperands(path string, value, values bool, collector *issueCollector) bool {
	valid := rejectValueMember(path, value, collector)
	return rejectValuesMember(path, values, collector) && valid
}

func rejectValueMember(path string, present bool, collector *issueCollector) bool {
	if present {
		collector.add(rawMemberPath(path, "value"), restqueryv1.IssueArityMismatch)
	}
	return !present
}

func rejectValuesMember(path string, present bool, collector *issueCollector) bool {
	if present {
		collector.add(rawMemberPath(path, "values"), restqueryv1.IssueArityMismatch)
	}
	return !present
}

func validateOperandCount(
	arity api.PredicateArity, count int, path string, collector *issueCollector,
) bool {
	valuesPath := rawMemberPath(path, "values")
	if arity == api.PredicateArityList && count > maximumWireOperandCount {
		collector.addBudget(valuesPath, restqueryv1.IssueOperandCountExceeded)
		return false
	}
	if arity == api.PredicateArityList && count == 0 ||
		arity == api.PredicateArityPair && count != 2 {
		collector.add(valuesPath, restqueryv1.IssueArityMismatch)
		return false
	}
	return true
}
