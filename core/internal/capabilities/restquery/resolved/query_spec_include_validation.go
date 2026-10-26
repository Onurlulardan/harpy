package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func validateIncludeTree(
	includes []Include,
	root api.CatalogReference,
	ancestors map[resourceNodeKey]struct{},
	depth int,
) (int, int, bool) {
	rootKey, valid := catalogResourceKey(root)
	if !valid {
		return 0, 0, false
	}
	ownedAncestors := cloneResourceNodes(ancestors)
	ownedAncestors[rootKey] = struct{}{}
	projections, edges := 0, 0
	for _, include := range includes {
		childProjections, childEdges, childValid := validateIncludeNode(
			include, ownedAncestors, depth,
		)
		if !childValid {
			return 0, 0, false
		}
		projections += childProjections
		edges += childEdges
	}
	return projections, edges, true
}

func validateIncludeNode(
	include Include, ancestors map[resourceNodeKey]struct{}, depth int,
) (int, int, bool) {
	if include.Validate() != nil || depth > 2 {
		return 0, 0, false
	}
	targetKey, valid := catalogResourceKey(include.target)
	if !valid {
		return 0, 0, false
	}
	if _, cycle := ancestors[targetKey]; cycle {
		return 0, 0, false
	}
	ownedAncestors := cloneResourceNodes(ancestors)
	ownedAncestors[targetKey] = struct{}{}
	projections := visibleProjectionCount(include.projections)
	edges := 1
	for _, child := range include.children {
		childProjections, childEdges, childValid := validateIncludeNode(
			child, ownedAncestors, depth+1,
		)
		if !childValid {
			return 0, 0, false
		}
		projections += childProjections
		edges += childEdges
	}
	return projections, edges, true
}

func catalogResourceKey(reference api.CatalogReference) (resourceNodeKey, bool) {
	owner, ownerErr := reference.Owner()
	resourceID, resourceErr := reference.ResourceID()
	if ownerErr != nil || resourceErr != nil {
		return resourceNodeKey{}, false
	}
	return resourceNodeKey{owner: owner, resourceID: resourceID}, true
}

func cloneResourceNodes(
	values map[resourceNodeKey]struct{},
) map[resourceNodeKey]struct{} {
	result := make(map[resourceNodeKey]struct{}, len(values)+1)
	for key := range values {
		result[key] = struct{}{}
	}
	return result
}
