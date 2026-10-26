package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveAggregateSorts(
	requests []api.RequestAggregateSort,
	groupOrdinals map[string]int,
	measureOrdinals map[string]int,
	groupCount int,
	collector *resolutionIssues,
) []resolved.AggregateSort {
	result := make([]resolved.AggregateSort, 0, len(requests)+groupCount)
	explicitGroups := make(map[int]struct{}, groupCount)
	for index, request := range requests {
		path := indexedPath("/aggregate/sort", index)
		reference, valid := resolveAggregateSortReference(
			request, path, groupOrdinals, measureOrdinals, collector,
		)
		if !valid {
			continue
		}
		sort, err := resolved.NewAggregateSort(reference, request.Direction(), false)
		if err != nil {
			collector.add(path, restqueryv1.IssueReferenceInvalid)
			continue
		}
		result = append(result, sort)
		if request.Reference() == api.AggregateReferenceGroup {
			explicitGroups[groupOrdinals[request.Name()]] = struct{}{}
		}
	}
	for ordinal := 0; ordinal < groupCount; ordinal++ {
		if _, found := explicitGroups[ordinal]; found {
			continue
		}
		reference, err := resolved.NewAggregateReference(
			api.AggregateReferenceGroup, ordinal, mustRequestPath("/aggregate/sort"),
		)
		if err != nil {
			collector.add("/aggregate/sort", restqueryv1.IssueReferenceInvalid)
			continue
		}
		sort, err := resolved.NewAggregateSort(
			reference, api.DirectionAscending, true,
		)
		if err != nil {
			collector.add("/aggregate/sort", restqueryv1.IssueReferenceInvalid)
			continue
		}
		result = append(result, sort)
	}
	return result
}

func resolveAggregateSortReference(
	request api.RequestAggregateSort,
	path string,
	groupOrdinals map[string]int,
	measureOrdinals map[string]int,
	collector *resolutionIssues,
) (resolved.AggregateReference, bool) {
	ordinals := groupOrdinals
	if request.Reference() == api.AggregateReferenceMeasure {
		ordinals = measureOrdinals
	}
	ordinal, found := ordinals[request.Name()]
	if !found {
		collector.add(path+"/name", restqueryv1.IssueFieldUnknown)
		return resolved.AggregateReference{}, false
	}
	reference, err := resolved.NewAggregateReference(
		request.Reference(), ordinal, mustRequestPath(path),
	)
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return resolved.AggregateReference{}, false
	}
	return reference, true
}
