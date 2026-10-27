package v1query

import (
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func decodePagination(
	document v1Document, policy manifest, collector *issueCollector,
) PagePlan {
	page := int64(1)
	valid := true
	if document.page != nil {
		parsed, err := strconv.ParseInt(document.page.value, 10, 64)
		if err != nil {
			collector.addPaginationParse(document.page, true)
			valid = false
		} else if parsed > 0 {
			page = parsed
		}
	}
	limit := policy.defaultLimit
	limitPath := "/limit"
	if document.limit != nil {
		limitPath = document.limit.path
		parsed, err := strconv.ParseInt(document.limit.value, 10, 64)
		if err != nil {
			collector.addPaginationParse(document.limit, false)
			valid = false
		} else if parsed <= 0 || parsed > maximumLegacyLimit {
			limit = maximumLegacyLimit
		} else {
			limit = int(parsed)
		}
	}
	if limit > 500 || limit > policy.maximumPageSize {
		collector.addBudget(limitPath, restqueryv1.IssuePageSizeExceeded)
		return PagePlan{}
	}
	if !valid {
		return PagePlan{}
	}
	plan, err := newPagePlan(page, limit)
	if err != nil {
		pagePath := "/page"
		if document.page != nil {
			pagePath = document.page.path
		}
		collector.addBudget(pagePath, restqueryv1.IssuePageOrdinalOverflow)
		return PagePlan{}
	}
	return plan
}

func (collector *issueCollector) addPaginationParse(
	member *documentMember, page bool,
) {
	if signedDecimal(member.value) {
		if page {
			collector.addBudget(member.path, restqueryv1.IssuePageOrdinalOverflow)
		} else {
			collector.addBudget(member.path, restqueryv1.IssuePageSizeExceeded)
		}
		return
	}
	collector.add(member.path, restqueryv1.IssueValueNotCanonical)
}

func signedDecimal(value string) bool {
	if value == "" {
		return false
	}
	start := 0
	if value[0] == '+' || value[0] == '-' {
		start = 1
	}
	if start == len(value) {
		return false
	}
	for _, character := range value[start:] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
