package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeAggregateFunction(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (api.AggregateFunction, bool) {
	if !present {
		return 0, false
	}
	value, valid := decodeString(raw, path, collector)
	if !valid {
		return 0, false
	}
	function, found := api.ParseAggregateFunction(value)
	if !found {
		collector.add(path, restqueryv1.IssueFunctionUnknown)
		return 0, false
	}
	return function, true
}

func decodeMeasureField(
	object rawObject,
	path string,
	function api.AggregateFunction,
	functionValid bool,
	collector *issueCollector,
) (*string, bool) {
	raw, present := object["field"]
	fieldPath := rawMemberPath(path, "field")
	if !functionValid {
		return nil, false
	}
	if function == api.AggregateFunctionCount {
		if present {
			collector.add(fieldPath, restqueryv1.IssueMemberForbidden)
			return nil, false
		}
		return nil, true
	}
	if !present {
		collector.add(fieldPath, restqueryv1.IssueMemberRequired)
		return nil, false
	}
	field, valid := decodeFieldName(raw, fieldPath, collector)
	if !valid {
		return nil, false
	}
	return &field, true
}
