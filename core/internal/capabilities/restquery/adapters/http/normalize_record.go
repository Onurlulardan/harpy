package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/contract"
)

func normalizeRecordDocument(
	document contract.QueryDocument, collector *issueCollector,
) (api.QueryRequest, bool) {
	selectRaw, selectPresent := memberValue(document.Select())
	selects, selectValid := decodeRecordSelect(selectRaw, selectPresent, collector)
	whereRaw, wherePresent := memberValue(document.Where())
	where, _, whereValid := decodeOptionalFilter(
		whereRaw, wherePresent, "/where", collector,
	)
	sortRaw, sortPresent := memberValue(document.Sort())
	sorts, sortValid := decodeSorts(sortRaw, sortPresent, "/sort", collector)
	pageRaw, pagePresent := memberValue(document.Page())
	page, pageValid := decodePage(pageRaw, pagePresent, "/page", collector)
	searchRaw, searchPresent := memberValue(document.Search())
	search, searchValid := decodeSearch(searchRaw, searchPresent, "/search", collector)
	countRaw, countPresent := memberValue(document.Count())
	count, countValid := decodeCount(countRaw, countPresent, "/count", collector)
	includeRaw, includePresent := memberValue(document.Include())
	includes, includeValid := decodeIncludes(includeRaw, includePresent, "/include", collector)
	valid := selectValid && whereValid && sortValid && pageValid && searchValid &&
		countValid && includeValid
	if valid {
		valid = validateRecordCombinedBudgets(selects, where, includes, count, collector)
	}
	if !valid || collector.hasIssues() {
		return api.QueryRequest{}, false
	}
	result, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: selects, Where: where, Sort: sorts, Page: page,
		Search: search, Count: count, Include: includes,
	})
	if err != nil {
		collector.add("", restqueryv1.IssueValueNotAllowed)
		return api.QueryRequest{}, false
	}
	return result, true
}

func decodeRecordSelect(
	raw json.RawMessage, present bool, collector *issueCollector,
) ([]string, bool) {
	if !present {
		return []string{}, true
	}
	return decodeFieldList(
		raw, "/select", 50, restqueryv1.IssueSelectCountExceeded, collector,
	)
}
