package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeIncludeSelect(
	object rawObject, path string, collector *issueCollector,
) ([]string, bool) {
	raw, present := object["select"]
	if !present {
		return []string{}, true
	}
	return decodeFieldList(
		raw, rawMemberPath(path, "select"), 50,
		restqueryv1.IssueSelectCountExceeded, collector,
	)
}

func decodeIncludeLimit(
	object rawObject, path string, collector *issueCollector,
) (*int, bool) {
	raw, present := object["limit"]
	if !present {
		return nil, true
	}
	limitPath := rawMemberPath(path, "limit")
	limit, valid := decodeInteger(raw, limitPath, collector)
	if !valid {
		return nil, false
	}
	if limit > 100 {
		collector.addBudget(limitPath, restqueryv1.IssueIncludeResultCountExceeded)
		return nil, false
	}
	if limit < 1 {
		collector.add(limitPath, restqueryv1.IssueValueOutOfRange)
		return nil, false
	}
	return &limit, true
}

func decodeIncludeChildren(
	object rawObject,
	path string,
	depth int,
	state *includeDecodeState,
	collector *issueCollector,
) ([]api.RequestInclude, bool) {
	raw, present := object["include"]
	if !present {
		return []api.RequestInclude{}, true
	}
	return decodeIncludeArray(
		json.RawMessage(raw), rawMemberPath(path, "include"), depth+1, state, collector,
	)
}
