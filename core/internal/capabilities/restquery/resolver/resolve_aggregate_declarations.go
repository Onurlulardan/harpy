package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveAggregateGroups(
	names []string, catalog api.FieldCatalog, collector *resolutionIssues,
) ([]resolved.AggregateGroup, map[string]int) {
	groups := make([]resolved.AggregateGroup, 0, len(names))
	ordinals := make(map[string]int, len(names))
	for index, name := range names {
		path := indexedPath("/aggregate/group_by", index)
		field, found, err := catalog.FieldByPublicName(name)
		if err != nil || !found {
			collector.add(path, restqueryv1.IssueFieldUnknown)
			continue
		}
		key, keyErr := resolved.NewOutputKey(name)
		group, groupErr := resolved.NewAggregateGroup(
			field, key, mustRequestPath(path),
		)
		if keyErr != nil || groupErr != nil {
			collector.add(path, restqueryv1.IssueCapabilityRequired)
			continue
		}
		ordinals[name] = len(groups)
		groups = append(groups, group)
	}
	return groups, ordinals
}

func resolveAggregateMeasures(
	requests []api.RequestMeasure,
	catalog api.FieldCatalog,
	collector *resolutionIssues,
) ([]resolved.AggregateMeasure, map[string]int) {
	measures := make([]resolved.AggregateMeasure, 0, len(requests))
	ordinals := make(map[string]int, len(requests))
	for index, request := range requests {
		path := indexedPath("/aggregate/measures", index)
		key, err := resolved.NewOutputKey(request.Name())
		if err != nil {
			collector.add(path+"/name", restqueryv1.IssueNameInvalid)
			continue
		}
		field, valid := resolveMeasureField(request, catalog, path, collector)
		if !valid {
			continue
		}
		measure, err := resolved.NewAggregateMeasure(
			key, request.Function(), field, mustRequestPath(path),
		)
		if err != nil {
			collector.add(path+"/function", restqueryv1.IssueFunctionNotAllowed)
			continue
		}
		ordinals[request.Name()] = len(measures)
		measures = append(measures, measure)
	}
	return measures, ordinals
}

func resolveMeasureField(
	request api.RequestMeasure,
	catalog api.FieldCatalog,
	path string,
	collector *resolutionIssues,
) (*api.FieldDefinition, bool) {
	name, hasField := request.Field()
	if !hasField {
		return nil, request.Function() == api.AggregateFunctionCount
	}
	field, found, err := catalog.FieldByPublicName(name)
	if err != nil || !found {
		collector.add(path+"/field", restqueryv1.IssueFieldUnknown)
		return nil, false
	}
	capabilities, err := field.Capabilities()
	if err != nil {
		collector.add(path+"/function", restqueryv1.IssueFunctionNotAllowed)
		return nil, false
	}
	allowed, err := capabilities.AllowsAggregate(request.Function())
	if err != nil || !allowed {
		collector.add(path+"/function", restqueryv1.IssueFunctionNotAllowed)
		return nil, false
	}
	return &field, true
}
