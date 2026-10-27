package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeDeclaredReference(
	kind string,
	raw json.RawMessage,
	path string,
	groups map[string]struct{},
	measures map[string]struct{},
	collector *issueCollector,
) (api.AggregateReference, string, bool) {
	memberPath := rawMemberPath(path, kind)
	name, valid := decodeReferenceName(kind, raw, memberPath, collector)
	if !valid {
		return 0, "", false
	}
	if kind == "group" {
		if _, declared := groups[name]; !declared {
			collector.add(memberPath, restqueryv1.IssueReferenceInvalid)
			return 0, "", false
		}
		return api.AggregateReferenceGroup, name, true
	}
	if _, declared := measures[name]; !declared {
		collector.add(memberPath, restqueryv1.IssueReferenceInvalid)
		return 0, "", false
	}
	return api.AggregateReferenceMeasure, name, true
}

func decodeReferenceName(
	kind string, raw []byte, path string, collector *issueCollector,
) (string, bool) {
	if kind == "group" {
		return decodeFieldName(raw, path, collector)
	}
	name, valid := decodeName(raw, path, collector)
	if !valid {
		return "", false
	}
	if !validMeasureWireName(name) {
		collector.add(path, restqueryv1.IssueNameInvalid)
		return "", false
	}
	return name, true
}
