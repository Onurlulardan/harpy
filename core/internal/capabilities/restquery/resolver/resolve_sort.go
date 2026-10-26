package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveSorts(
	catalog api.FieldCatalog,
	requests []api.RequestSort,
	stableID api.FieldDefinition,
	basePath string,
	requireStable bool,
	collector *resolutionIssues,
) []resolved.Sort {
	stableIdentifier := fieldUUID(stableID)
	result := make([]resolved.Sort, 0, len(requests)+1)
	var stableSort *resolved.Sort
	for index, request := range requests {
		path := indexedPath(basePath, index)
		field, found, err := catalog.FieldByPublicName(request.Field())
		if err != nil || !found {
			collector.add(path+"/field", restqueryv1.IssueFieldUnknown)
			continue
		}
		sort, err := resolved.NewSort(
			field, request.Direction(), mustRequestPath(path), false,
		)
		if err != nil {
			collector.add(path+"/field", restqueryv1.IssueCapabilityRequired)
			continue
		}
		if fieldUUID(field) == stableIdentifier && requireStable {
			owned := sort
			stableSort = &owned
			continue
		}
		result = append(result, sort)
	}
	if !requireStable {
		return result
	}
	if stableSort != nil {
		return append(result, *stableSort)
	}
	sort, err := resolved.NewSort(
		stableID, api.DirectionAscending, mustRequestPath(basePath), true,
	)
	if err != nil {
		collector.add(basePath, restqueryv1.IssueCapabilityRequired)
		return result
	}
	return append(result, sort)
}
