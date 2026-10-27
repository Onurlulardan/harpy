package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func normalizeGETDocument(
	document getDocument, collector *issueCollector,
) (api.QueryRequest, bool) {
	selects, selectValid := decodeGETSelects(document, collector)
	where, filterValid := decodeGETFilters(document, collector)
	sorts, sortValid := decodeGETSorts(document, collector)
	search, searchValid := decodeGETSearch(document, collector)
	page, pageValid := decodeGETPage(document, collector)
	count, countValid := decodeGETCount(document, collector)
	valid := selectValid && filterValid && sortValid && searchValid &&
		pageValid && countValid
	if valid {
		valid = validateRecordCombinedBudgets(
			selects, where, []api.RequestInclude{}, count, collector,
		)
	}
	if !valid || collector.hasIssues() {
		return api.QueryRequest{}, false
	}
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: selects, Where: where, Sort: sorts, Page: page,
		Search: search, Count: count, Include: []api.RequestInclude{},
	})
	if err != nil {
		collector.add("", restqueryv1.IssueValueNotAllowed)
		return api.QueryRequest{}, false
	}
	return request, true
}
