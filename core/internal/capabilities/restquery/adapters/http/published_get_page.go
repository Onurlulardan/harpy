package httpquery

import (
	"strconv"
	"strings"
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

func normalizePublishedGETPage(
	document publishedGETDocument,
	collector *publishedIssueCollector,
) *published.PageOverride {
	if document.pageSize == nil && document.pageAfter == nil {
		return nil
	}
	var size *int
	if document.pageSize != nil {
		parsed, valid := decodePublishedGETPageSize(*document.pageSize, collector)
		if valid {
			size = &parsed
		}
	}
	var after *string
	if document.pageAfter != nil {
		value := document.pageAfter.value
		if value == "" || !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') {
			collector.add(document.pageAfter.pointer, restqueryv1.IssueValueNotAllowed)
		} else if len(value) > 16*1024 {
			collector.add(document.pageAfter.pointer, restqueryv1.IssueCursorSizeExceeded)
		} else {
			after = &value
		}
	}
	if collector.hasIssues() {
		return nil
	}
	page, err := published.NewPageOverride(size, after)
	if err != nil {
		collector.add("/page", restqueryv1.IssueValueNotAllowed)
		return nil
	}
	return &page
}

func decodePublishedGETPageSize(
	member getMember, collector *publishedIssueCollector,
) (int, bool) {
	if !canonicalGETDecimal(member.value) {
		collector.add(member.pointer, restqueryv1.IssueValueNotCanonical)
		return 0, false
	}
	parsed, err := strconv.ParseUint(member.value, 10, strconv.IntSize)
	if err != nil {
		collector.add(member.pointer, restqueryv1.IssueValueOutOfRange)
		return 0, false
	}
	size := int(parsed)
	if size < 1 {
		collector.add(member.pointer, restqueryv1.IssueValueOutOfRange)
		return 0, false
	}
	if size > 500 {
		collector.add(member.pointer, restqueryv1.IssuePageSizeExceeded)
		return 0, false
	}
	return size, true
}
