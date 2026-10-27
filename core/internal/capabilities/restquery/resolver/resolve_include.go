package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type includeResourceKey struct {
	owner      string
	resourceID string
}

func resolveIncludes(
	requests []api.RequestInclude,
	root api.FieldCatalog,
	catalogs api.CatalogSet,
	collector *resolutionIssues,
) ([]resolved.Include, int) {
	rootReference, err := root.Reference()
	if err != nil {
		collector.add("/include", restqueryv1.IssueReferenceInvalid)
		return nil, 0
	}
	ancestors := map[includeResourceKey]struct{}{includeCatalogKey(rootReference): {}}
	return resolveIncludeLevel(
		requests, root, catalogs, "/include", ancestors, collector,
	)
}

func resolveIncludeLevel(
	requests []api.RequestInclude,
	source api.FieldCatalog,
	catalogs api.CatalogSet,
	basePath string,
	ancestors map[includeResourceKey]struct{},
	collector *resolutionIssues,
) ([]resolved.Include, int) {
	result := make([]resolved.Include, 0, len(requests))
	projectionCount := 0
	for index, request := range requests {
		path := indexedPath(basePath, index)
		include, count, valid := resolveInclude(
			request, source, catalogs, path, ancestors, collector,
		)
		if valid {
			result = append(result, include)
			projectionCount += count
		}
	}
	return result, projectionCount
}

func resolveInclude(
	request api.RequestInclude,
	source api.FieldCatalog,
	catalogs api.CatalogSet,
	path string,
	ancestors map[includeResourceKey]struct{},
	collector *resolutionIssues,
) (resolved.Include, int, bool) {
	relation, found, err := source.RelationByPublicName(request.Relation())
	if err != nil || !found {
		collector.add(path+"/relation", restqueryv1.IssueRelationUnknown)
		return resolved.Include{}, 0, false
	}
	enabled, err := relation.IncludeEnabled()
	if err != nil || !enabled {
		collector.add(path+"/relation", restqueryv1.IssueCapabilityRequired)
		return resolved.Include{}, 0, false
	}
	targetReference, err := relation.Target()
	if err != nil {
		collector.add(path+"/relation", restqueryv1.IssueReferenceInvalid)
		return resolved.Include{}, 0, false
	}
	targetKey := includeCatalogKey(targetReference)
	if _, cycle := ancestors[targetKey]; cycle {
		collector.add(path+"/relation", restqueryv1.IssueRelationCycle)
		return resolved.Include{}, 0, false
	}
	target, found, err := catalogs.Catalog(targetReference)
	if err != nil || !found {
		collector.add(path+"/relation", restqueryv1.IssueRelationUnknown)
		return resolved.Include{}, 0, false
	}
	return resolveKnownInclude(
		request, relation, target, catalogs, path, ancestors, collector,
	)
}

func includeCatalogKey(reference api.CatalogReference) includeResourceKey {
	owner, _ := reference.Owner()
	resourceID, _ := reference.ResourceID()
	return includeResourceKey{owner: owner, resourceID: resourceID.String()}
}

func cloneIncludeAncestors(
	values map[includeResourceKey]struct{}, next includeResourceKey,
) map[includeResourceKey]struct{} {
	result := make(map[includeResourceKey]struct{}, len(values)+1)
	for key := range values {
		result[key] = struct{}{}
	}
	result[next] = struct{}{}
	return result
}
