package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveRecord(
	request api.QueryRequest,
	root api.FieldCatalog,
	catalogs api.CatalogSet,
	collector *resolutionIssues,
) resolved.RecordQuery {
	projections, stableID := resolveProjections(
		root, request.Select(), "/select", false, true, collector,
	)
	var where *resolved.Filter
	if requestWhere, found := request.Where(); found {
		resolvedWhere, valid := resolveFilter(root, requestWhere, "/where", collector)
		if valid {
			where = &resolvedWhere
		}
	}
	sorts := resolveSorts(root, request.Sort(), stableID, "/sort", true, collector)
	search := []resolved.SearchTerm{}
	if requestSearch, found := request.Search(); found {
		search = resolveSearch(root, requestSearch, "/search", collector)
	}
	included, includedProjectionCount := resolveIncludes(
		request.Includes(), root, catalogs, collector,
	)
	if visibleResolvedProjections(projections)+includedProjectionCount > 50 {
		collector.budget("", restqueryv1.IssueSelectCountExceeded)
	}
	exactCountEnabled := resolveExactCountCapability(request, root, collector)
	after, hasAfter := request.Page().After()
	var afterPointer *string
	if hasAfter {
		afterPointer = &after
	}
	page, err := resolved.NewPage(request.Page().Size(), afterPointer)
	if err != nil {
		collector.add("/page", restqueryv1.IssueReferenceInvalid)
	}
	record, err := resolved.NewRecordQuery(
		projections, where, sorts, search, page, request.Count(),
		mustRequestPath("/count"), exactCountEnabled, included,
	)
	if err != nil && !collector.hasIssues() {
		collector.add("", restqueryv1.IssueReferenceInvalid)
	}
	return record
}

func visibleResolvedProjections(projections []resolved.Projection) int {
	count := 0
	for _, projection := range projections {
		hidden, err := projection.Hidden()
		if err == nil && !hidden {
			count++
		}
	}
	return count
}
