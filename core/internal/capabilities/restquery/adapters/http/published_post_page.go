package httpquery

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

func decodePublishedPOSTPage(
	root publishedPOSTRoot, collector *publishedIssueCollector,
) *published.PageOverride {
	raw, present := root["page"]
	if !present {
		return nil
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		collector.add("/page", restqueryv1.IssueTypeMismatch)
		return nil
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(trimmed, &object) != nil || object == nil {
		collector.add("/page", restqueryv1.IssueTypeMismatch)
		return nil
	}
	for name := range object {
		if name != "size" && name != "after" {
			collector.add(
				appendJSONPointer("/page", name), restqueryv1.IssueParameterUnknown,
			)
		}
	}
	sizeRaw, sizePresent := object["size"]
	afterRaw, afterPresent := object["after"]
	if !sizePresent {
		collector.add("/page/size", restqueryv1.IssueParameterRequired)
	}
	if !afterPresent {
		collector.add("/page/after", restqueryv1.IssueParameterRequired)
	}
	if !sizePresent || !afterPresent {
		return nil
	}
	size, sizeValid := decodePublishedPOSTPageSize(sizeRaw, collector)
	after, afterValid := decodePublishedPOSTAfter(afterRaw, collector)
	if !sizeValid || !afterValid || collector.hasIssues() {
		return nil
	}
	page, err := published.NewPageOverride(&size, after)
	if err != nil {
		collector.add("/page", restqueryv1.IssueValueNotAllowed)
		return nil
	}
	return &page
}

func decodePublishedPOSTPageSize(
	raw json.RawMessage, collector *publishedIssueCollector,
) (int, bool) {
	value := string(bytes.TrimSpace(raw))
	if value == "" || strings.ContainsAny(value, ".eE") ||
		value[0] != '-' && (value[0] < '0' || value[0] > '9') {
		collector.add("/page/size", restqueryv1.IssueTypeMismatch)
		return 0, false
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || int64(int(parsed)) != parsed {
		collector.add("/page/size", restqueryv1.IssueValueOutOfRange)
		return 0, false
	}
	size := int(parsed)
	if size < 1 {
		collector.add("/page/size", restqueryv1.IssueValueOutOfRange)
		return 0, false
	}
	if size > 500 {
		collector.add("/page/size", restqueryv1.IssuePageSizeExceeded)
		return 0, false
	}
	return size, true
}

func decodePublishedPOSTAfter(
	raw json.RawMessage, collector *publishedIssueCollector,
) (*string, bool) {
	if rawIsNull(raw) {
		return nil, true
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		collector.add("/page/after", restqueryv1.IssueTypeMismatch)
		return nil, false
	}
	if value == "" || !utf8.ValidString(value) {
		collector.add("/page/after", restqueryv1.IssueValueNotAllowed)
		return nil, false
	}
	if strings.ContainsRune(value, '\x00') {
		collector.add("/page/after", restqueryv1.IssueNULForbidden)
		return nil, false
	}
	if len(value) > 16*1024 {
		collector.add("/page/after", restqueryv1.IssueCursorSizeExceeded)
		return nil, false
	}
	return &value, true
}
