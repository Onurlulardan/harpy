package httpquery

import (
	"bytes"
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

type rawObject map[string]json.RawMessage

func decodeObject(
	raw []byte, path string, allowed []string, collector *issueCollector,
) (rawObject, bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		collector.add(path, restqueryv1.IssueTypeMismatch)
		return nil, false
	}
	var object rawObject
	if err := json.Unmarshal(trimmed, &object); err != nil || object == nil {
		collector.add(path, restqueryv1.IssueTypeMismatch)
		return nil, false
	}
	for name := range object {
		if !containsMember(allowed, name) {
			collector.add(appendJSONPointer(path, name), restqueryv1.IssueMemberUnknown)
		}
	}
	return object, true
}

func containsMember(allowed []string, candidate string) bool {
	for _, name := range allowed {
		if name == candidate {
			return true
		}
	}
	return false
}

func rawIsNull(raw []byte) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func rawMemberPath(path, member string) string {
	return appendJSONPointer(path, member)
}

func requireMember(
	object rawObject, path string, member string, collector *issueCollector,
) (json.RawMessage, bool) {
	raw, found := object[member]
	if !found {
		collector.add(rawMemberPath(path, member), restqueryv1.IssueMemberRequired)
	}
	return raw, found
}
