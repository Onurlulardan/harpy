package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodePredicateFilter(
	object rawObject, path string, collector *issueCollector,
) (decodedFilter, bool) {
	fieldRaw, fieldPresent := requireMember(object, path, "field", collector)
	opRaw, operatorPresent := requireMember(object, path, "op", collector)
	field, fieldValid := decodeFieldNameIfPresent(fieldRaw, fieldPresent, rawMemberPath(path, "field"), collector)
	operator, operatorValid := decodeOperatorIfPresent(
		opRaw, operatorPresent, rawMemberPath(path, "op"), collector,
	)
	values, valuesValid := decodePredicateOperands(object, path, operator, operatorValid, collector)
	metrics := predicateMetrics(len(values))
	if !validateFilterMetrics(metrics, path, collector) {
		valuesValid = false
	}
	if !fieldValid || !operatorValid || !valuesValid {
		return decodedFilter{metrics: metrics}, false
	}
	filter, err := api.NewRequestPredicate(field, operator, values)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return decodedFilter{metrics: metrics}, false
	}
	return decodedFilter{filter: filter, metrics: metrics}, true
}

func decodeFieldNameIfPresent(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (string, bool) {
	if !present {
		return "", false
	}
	return decodeFieldName(raw, path, collector)
}

func decodeOperatorIfPresent(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (api.Operator, bool) {
	if !present {
		return 0, false
	}
	value, valid := decodeString(raw, path, collector)
	if !valid {
		return 0, false
	}
	operator, found := api.ParseOperator(value)
	if !found {
		collector.add(path, restqueryv1.IssueOperatorUnknown)
		return 0, false
	}
	return operator, true
}
