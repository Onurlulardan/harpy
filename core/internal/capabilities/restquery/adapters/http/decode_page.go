package httpquery

import (
	"encoding/json"
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var pageMembers = []string{"size", "after"}

func decodePage(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (api.RequestPage, bool) {
	if !present {
		page, _ := api.NewRequestPage(50, nil)
		return page, true
	}
	start := len(collector.issues)
	object, valid := decodeObject(raw, path, pageMembers, collector)
	if !valid {
		return api.RequestPage{}, false
	}
	sizeRaw, sizePresent := requireMember(object, path, "size", collector)
	afterRaw, afterPresent := requireMember(object, path, "after", collector)
	size, sizeValid := 0, false
	if sizePresent {
		size, sizeValid = decodeInteger(sizeRaw, rawMemberPath(path, "size"), collector)
		if sizeValid {
			sizeValid = validatePageSize(size, rawMemberPath(path, "size"), collector)
		}
	}
	after, afterValid := decodeAfter(afterRaw, afterPresent, rawMemberPath(path, "after"), collector)
	if !sizeValid || !afterValid || len(collector.issues) != start {
		return api.RequestPage{}, false
	}
	page, err := api.NewRequestPage(size, after)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return api.RequestPage{}, false
	}
	return page, true
}

func validatePageSize(size int, path string, collector *issueCollector) bool {
	if size > 500 {
		collector.addBudget(path, restqueryv1.IssuePageSizeExceeded)
		return false
	}
	if size < 1 {
		collector.add(path, restqueryv1.IssueValueOutOfRange)
		return false
	}
	return true
}

func decodeAfter(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (*string, bool) {
	if !present {
		return nil, false
	}
	if rawIsNull(raw) {
		return nil, true
	}
	value, valid := decodeString(raw, path, collector)
	if !valid {
		return nil, false
	}
	if value == "" || !utf8.ValidString(value) {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return nil, false
	}
	if len(value) > 16*1024 {
		collector.addBudget(path, restqueryv1.IssueCursorSizeExceeded)
		return nil, false
	}
	return &value, true
}
