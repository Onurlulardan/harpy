package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var filterMembers = []string{"and", "or", "not", "field", "op", "value", "values"}

type decodedFilter struct {
	filter  api.RequestFilter
	metrics filterMetrics
}

func decodeOptionalFilter(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (*api.RequestFilter, filterMetrics, bool) {
	if !present || rawIsNull(raw) {
		return nil, filterMetrics{}, true
	}
	decoded, valid := decodeFilter(raw, path, collector)
	if !valid {
		return nil, decoded.metrics, false
	}
	result := decoded.filter
	return &result, decoded.metrics, true
}

func decodeFilter(
	raw []byte, path string, collector *issueCollector,
) (decodedFilter, bool) {
	start := len(collector.issues)
	object, objectValid := decodeObject(raw, path, filterMembers, collector)
	if !objectValid {
		return decodedFilter{}, false
	}
	booleanNames := presentMembers(object, []string{"and", "or", "not"})
	predicateNames := presentMembers(object, []string{"field", "op", "value", "values"})
	if len(booleanNames) == 0 {
		decoded, valid := decodePredicateFilter(object, path, collector)
		return decoded, valid && len(collector.issues) == start
	}
	selected := booleanNames[0]
	for _, name := range booleanNames[1:] {
		collector.add(rawMemberPath(path, name), restqueryv1.IssueMemberForbidden)
	}
	for _, name := range predicateNames {
		collector.add(rawMemberPath(path, name), restqueryv1.IssueMemberForbidden)
	}
	decoded, valid := decodeBooleanFilter(selected, object[selected], path, collector)
	return decoded, valid && len(collector.issues) == start
}

func presentMembers(object rawObject, names []string) []string {
	result := make([]string, 0, len(names))
	for _, name := range names {
		if _, present := object[name]; present {
			result = append(result, name)
		}
	}
	return result
}
