package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var havingMembers = []string{"and", "or", "not", "group", "measure", "op", "value", "values"}

type decodedHaving struct {
	having  api.RequestHaving
	metrics havingMetrics
}

func decodeOptionalHaving(
	raw json.RawMessage,
	present bool,
	path string,
	groups map[string]struct{},
	measures map[string]struct{},
	collector *issueCollector,
) (*api.RequestHaving, havingMetrics, bool) {
	if !present || rawIsNull(raw) {
		return nil, havingMetrics{}, true
	}
	decoded, valid := decodeHaving(raw, path, groups, measures, collector)
	if !valid {
		return nil, decoded.metrics, false
	}
	result := decoded.having
	return &result, decoded.metrics, true
}

func decodeHaving(
	raw []byte,
	path string,
	groups map[string]struct{},
	measures map[string]struct{},
	collector *issueCollector,
) (decodedHaving, bool) {
	start := len(collector.issues)
	object, objectValid := decodeObject(raw, path, havingMembers, collector)
	if !objectValid {
		return decodedHaving{}, false
	}
	booleanNames := presentMembers(object, []string{"and", "or", "not"})
	predicateNames := presentMembers(object, []string{"group", "measure", "op", "value", "values"})
	if len(booleanNames) == 0 {
		decoded, valid := decodeHavingPredicate(object, path, groups, measures, collector)
		return decoded, valid && len(collector.issues) == start
	}
	selected := booleanNames[0]
	for _, name := range booleanNames[1:] {
		collector.add(rawMemberPath(path, name), restqueryv1.IssueMemberForbidden)
	}
	for _, name := range predicateNames {
		collector.add(rawMemberPath(path, name), restqueryv1.IssueMemberForbidden)
	}
	decoded, valid := decodeHavingBoolean(
		selected, object[selected], path, groups, measures, collector,
	)
	return decoded, valid && len(collector.issues) == start
}
