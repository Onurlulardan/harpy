package httpquery

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func decodeString(
	raw []byte, path string, collector *issueCollector,
) (string, bool) {
	var value string
	if len(bytes.TrimSpace(raw)) == 0 || json.Unmarshal(raw, &value) != nil || rawIsNull(raw) {
		collector.add(path, restqueryv1.IssueTypeMismatch)
		return "", false
	}
	return value, true
}

func decodeInteger(
	raw []byte, path string, collector *issueCollector,
) (int, bool) {
	trimmed := string(bytes.TrimSpace(raw))
	if trimmed == "" || strings.ContainsAny(trimmed, ".eE") {
		collector.add(path, restqueryv1.IssueTypeMismatch)
		return 0, false
	}
	first := trimmed[0]
	if first != '-' && (first < '0' || first > '9') {
		collector.add(path, restqueryv1.IssueTypeMismatch)
		return 0, false
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || int64(int(parsed)) != parsed {
		collector.add(path, restqueryv1.IssueValueOutOfRange)
		return 0, false
	}
	return int(parsed), true
}

func decodeArray(
	raw []byte, path string, collector *issueCollector,
) ([]json.RawMessage, bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		collector.add(path, restqueryv1.IssueTypeMismatch)
		return nil, false
	}
	var values []json.RawMessage
	if err := json.Unmarshal(trimmed, &values); err != nil || values == nil {
		collector.add(path, restqueryv1.IssueTypeMismatch)
		return nil, false
	}
	return values, true
}
