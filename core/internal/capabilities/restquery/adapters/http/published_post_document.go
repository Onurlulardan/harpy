package httpquery

import (
	"bytes"
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

type publishedPOSTRoot map[string]json.RawMessage

func decodePublishedPOSTRoot(
	document []byte, collector *publishedIssueCollector,
) (publishedPOSTRoot, bool) {
	trimmed := bytes.TrimSpace(document)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		collector.add("", restqueryv1.IssueTypeMismatch)
		return nil, false
	}
	var root publishedPOSTRoot
	if json.Unmarshal(trimmed, &root) != nil || root == nil {
		collector.add("", restqueryv1.IssueTypeMismatch)
		return nil, false
	}
	for name := range root {
		if name != "params" && name != "page" {
			collector.add(
				appendJSONPointer("", name), restqueryv1.IssueParameterUnknown,
			)
		}
	}
	if _, found := root["params"]; !found {
		collector.add("/params", restqueryv1.IssueParameterRequired)
	}
	return root, true
}
