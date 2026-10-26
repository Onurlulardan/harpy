package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveAggregate(
	request api.QueryRequest,
	root api.FieldCatalog,
	collector *resolutionIssues,
) resolved.AggregateQuery {
	declaration, found := request.Aggregate()
	if !found {
		collector.add("/aggregate", restqueryv1.IssueMemberRequired)
		return resolved.AggregateQuery{}
	}
	var where *resolved.Filter
	if requestWhere, found := request.Where(); found {
		resolvedWhere, valid := resolveFilter(root, requestWhere, "/where", collector)
		if valid {
			where = &resolvedWhere
		}
	}
	search := []resolved.SearchTerm{}
	if requestSearch, found := request.Search(); found {
		search = resolveSearch(root, requestSearch, "/search", collector)
	}
	groups, groupOrdinals := resolveAggregateGroups(
		declaration.GroupBy(), root, collector,
	)
	measures, measureOrdinals := resolveAggregateMeasures(
		declaration.Measures(), root, collector,
	)
	if collector.hasIssues() {
		return resolved.AggregateQuery{}
	}
	var having *resolved.Having
	if requestHaving, found := declaration.Having(); found {
		resolvedHaving, valid := resolveHaving(
			requestHaving, "/aggregate/having", groups, measures,
			groupOrdinals, measureOrdinals, collector,
		)
		if valid {
			having = &resolvedHaving
		}
	}
	sorts := resolveAggregateSorts(
		declaration.Sort(), groupOrdinals, measureOrdinals, len(groups), collector,
	)
	if collector.hasIssues() {
		return resolved.AggregateQuery{}
	}
	query, err := resolved.NewAggregateQuery(
		groups, measures, where, search, having, sorts, declaration.Limit(),
	)
	if err != nil {
		collector.add("/aggregate", restqueryv1.IssueReferenceInvalid)
	}
	return query
}
