package httpquery

import (
	"strconv"
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (document *getDocument) addGETScalar(
	leaf getLeaf, collector *issueCollector,
) {
	root := leaf.segments[0]
	if root == "count" {
		if len(leaf.segments) != 1 {
			collector.add(leaf.pointer, restqueryv1.IssueValueNotCanonical)
			return
		}
		document.count = memberFromGETLeaf(leaf)
		return
	}
	if len(leaf.segments) != 2 {
		collector.add(leaf.pointer, restqueryv1.IssueValueNotCanonical)
		return
	}
	member := memberFromGETLeaf(leaf)
	if root == "search" && leaf.segments[1] == "term" {
		document.search = member
		return
	}
	if root == "page" {
		switch leaf.segments[1] {
		case "size":
			document.pageSize = member
		case "after":
			document.pageAfter = member
		default:
			collector.add(leaf.pointer, restqueryv1.IssueMemberUnknown)
		}
		return
	}
	collector.add(leaf.pointer, restqueryv1.IssueMemberUnknown)
}

func decodeGETSearch(
	document getDocument, collector *issueCollector,
) (*api.RequestSearch, bool) {
	if document.search == nil {
		return nil, true
	}
	term, valid := validateGETSearchTerm(*document.search, collector)
	if !valid {
		return nil, false
	}
	search, err := api.NewRequestSearch(term)
	if err != nil {
		collector.add(document.search.pointer, restqueryv1.IssueValueNotAllowed)
		return nil, false
	}
	return &search, true
}

func decodeGETPage(
	document getDocument, collector *issueCollector,
) (api.RequestPage, bool) {
	size, valid := 50, true
	if document.pageSize != nil {
		size, valid = decodeGETPageSize(*document.pageSize, collector)
	}
	var after *string
	if document.pageAfter != nil {
		value := document.pageAfter.value
		if value == "" || !utf8.ValidString(value) {
			collector.add(document.pageAfter.pointer, restqueryv1.IssueValueNotAllowed)
			valid = false
		} else if len(value) > 16*1024 {
			collector.addBudget(document.pageAfter.pointer, restqueryv1.IssueCursorSizeExceeded)
			valid = false
		} else {
			after = &value
		}
	}
	if !valid {
		return api.RequestPage{}, false
	}
	page, err := api.NewRequestPage(size, after)
	if err != nil {
		collector.add("/page", restqueryv1.IssueValueNotAllowed)
		return api.RequestPage{}, false
	}
	return page, true
}

func decodeGETPageSize(
	member getMember, collector *issueCollector,
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
	return size, validatePageSize(size, member.pointer, collector)
}

func canonicalGETDecimal(value string) bool {
	if value == "" || len(value) > 1 && value[0] == '0' {
		return false
	}
	for index := range len(value) {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

func decodeGETCount(
	document getDocument, collector *issueCollector,
) (api.CountMode, bool) {
	if document.count == nil {
		return api.CountModeNone, true
	}
	count, found := api.ParseCountMode(document.count.value)
	if !found {
		collector.add(document.count.pointer, restqueryv1.IssueValueNotAllowed)
		return 0, false
	}
	return count, true
}
