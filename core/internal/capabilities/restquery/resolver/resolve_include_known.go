package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveKnownInclude(
	request api.RequestInclude,
	relation api.RelationDefinition,
	target api.FieldCatalog,
	catalogs api.CatalogSet,
	path string,
	ancestors map[includeResourceKey]struct{},
	collector *resolutionIssues,
) (resolved.Include, int, bool) {
	reference, targetStorage, stableID, err := catalogRootParts(target)
	if err != nil {
		collector.add(path+"/relation", restqueryv1.IssueReferenceInvalid)
		return resolved.Include{}, 0, false
	}
	projections, _ := resolveProjections(
		target, request.Select(), path+"/select", true, true, collector,
	)
	var where *resolved.Filter
	if requestWhere, found := request.Where(); found {
		resolvedWhere, valid := resolveFilter(
			target, requestWhere, path+"/where", collector,
		)
		if valid {
			where = &resolvedWhere
		}
	}
	cardinality, err := relation.Cardinality()
	if err != nil {
		collector.add(path+"/relation", restqueryv1.IssueReferenceInvalid)
		return resolved.Include{}, 0, false
	}
	requireStable := cardinality == api.RelationCardinalityToMany
	sorts := resolveSorts(
		target, request.Sort(), stableID, path+"/sort", requireStable, collector,
	)
	limit, validLimit := resolveIncludeLimit(request, relation)
	if !validLimit {
		collector.add(path+"/limit", restqueryv1.IssueIncludeResultCountExceeded)
		return resolved.Include{}, 0, false
	}
	children, childProjectionCount := resolveIncludeLevel(
		request.Includes(), target, catalogs, path+"/include",
		cloneIncludeAncestors(ancestors, includeCatalogKey(reference)), collector,
	)
	relationStorage, err := relation.Storage()
	if err != nil {
		collector.add(path+"/relation", restqueryv1.IssueReferenceInvalid)
		return resolved.Include{}, 0, false
	}
	include, err := resolved.NewInclude(
		relation, reference, targetStorage, relationStorage,
		stableID, mustRequestPath(path),
		projections, where, sorts, limit, children,
	)
	if err != nil {
		if !collector.hasIssues() {
			collector.add(path, restqueryv1.IssueReferenceInvalid)
		}
		return resolved.Include{}, 0, false
	}
	return include, visibleResolvedProjections(projections) + childProjectionCount, true
}

func resolveIncludeLimit(
	request api.RequestInclude, relation api.RelationDefinition,
) (int, bool) {
	cardinality, err := relation.Cardinality()
	if err != nil {
		return 0, false
	}
	maximum, err := relation.MaximumIncludedRecords()
	if err != nil {
		return 0, false
	}
	if requested, found := request.Limit(); found {
		if requested < 1 || requested > maximum ||
			cardinality == api.RelationCardinalityToOne && requested != 1 {
			return 0, false
		}
		return requested, true
	}
	if cardinality == api.RelationCardinalityToOne {
		return 1, true
	}
	return min(50, maximum), true
}
