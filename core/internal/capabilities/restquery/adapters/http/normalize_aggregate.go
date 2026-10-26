package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/contract"
)

func normalizeAggregateDocument(
	document contract.QueryDocument,
	aggregateRaw json.RawMessage,
	collector *issueCollector,
) (api.QueryRequest, bool) {
	forbiddenAggregateRootMembers(document, collector)
	whereRaw, wherePresent := memberValue(document.Where())
	where, whereMetrics, whereValid := decodeOptionalFilter(
		whereRaw, wherePresent, "/where", collector,
	)
	searchRaw, searchPresent := memberValue(document.Search())
	search, searchValid := decodeSearch(searchRaw, searchPresent, "/search", collector)
	aggregate, havingMetrics, aggregateValid := decodeAggregate(
		aggregateRaw, "/aggregate", collector,
	)
	combinedBudgetValid := validateAggregateCombinedBudget(
		whereMetrics, havingMetrics, collector,
	)
	if !whereValid || !searchValid || !aggregateValid || !combinedBudgetValid ||
		collector.hasIssues() {
		return api.QueryRequest{}, false
	}
	result, err := api.NewAggregateQueryRequest(api.AggregateQueryRequestInput{
		Where: where, Search: search, Aggregate: aggregate,
	})
	if err != nil {
		collector.add("/aggregate", restqueryv1.IssueValueNotAllowed)
		return api.QueryRequest{}, false
	}
	return result, true
}

func forbiddenAggregateRootMembers(
	document contract.QueryDocument, collector *issueCollector,
) {
	members := []struct {
		name   string
		member contract.RawMember
	}{
		{"select", document.Select()}, {"sort", document.Sort()},
		{"page", document.Page()}, {"include", document.Include()},
		{"count", document.Count()},
	}
	for _, candidate := range members {
		if _, present := candidate.member.Value(); present {
			collector.add(
				rawMemberPath("", candidate.name), restqueryv1.IssueMemberForbidden,
			)
		}
	}
}
