package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var aggregateMembers = []string{"group_by", "measures", "having", "sort", "limit"}

func decodeAggregate(
	raw []byte, path string, collector *issueCollector,
) (api.RequestAggregate, havingMetrics, bool) {
	start := len(collector.issues)
	object, objectValid := decodeObject(raw, path, aggregateMembers, collector)
	if !objectValid {
		return api.RequestAggregate{}, havingMetrics{}, false
	}
	groups, groupsValid := decodeAggregateGroups(object, path, collector)
	measuresRaw, measuresPresent := object["measures"]
	measures, measuresValid := decodeMeasures(
		measuresRaw, measuresPresent, rawMemberPath(path, "measures"), collector,
	)
	groupSet, measureSet := aggregateNameSets(groups, measures)
	having, havingStats, havingValid := decodeOptionalHaving(
		object["having"], object["having"] != nil, rawMemberPath(path, "having"),
		groupSet, measureSet, collector,
	)
	sorts, sortsValid := decodeAggregateSorts(
		object["sort"], object["sort"] != nil, rawMemberPath(path, "sort"),
		groupSet, measureSet, collector,
	)
	limit, limitValid := decodeAggregateLimit(object, path, collector)
	valid := groupsValid && measuresValid && havingValid && sortsValid &&
		limitValid && len(collector.issues) == start
	if !valid {
		return api.RequestAggregate{}, havingStats, false
	}
	result, err := api.NewRequestAggregate(api.RequestAggregateInput{
		GroupBy: groups, Measures: measures, Having: having, Sort: sorts, Limit: limit,
	})
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return api.RequestAggregate{}, havingStats, false
	}
	return result, havingStats, true
}

func decodeAggregateGroups(
	object rawObject, path string, collector *issueCollector,
) ([]string, bool) {
	raw, present := object["group_by"]
	if !present {
		return []string{}, true
	}
	return decodeFieldList(
		raw, rawMemberPath(path, "group_by"), 3,
		restqueryv1.IssueGroupCountExceeded, collector,
	)
}

func decodeAggregateLimit(
	object rawObject, path string, collector *issueCollector,
) (*int, bool) {
	raw, present := object["limit"]
	if !present {
		return nil, true
	}
	limitPath := rawMemberPath(path, "limit")
	limit, valid := decodeInteger(raw, limitPath, collector)
	if !valid {
		return nil, false
	}
	if limit > 500 {
		collector.addBudget(limitPath, restqueryv1.IssueResultCountExceeded)
		return nil, false
	}
	if limit < 1 {
		collector.add(limitPath, restqueryv1.IssueValueOutOfRange)
		return nil, false
	}
	return &limit, true
}

func aggregateNameSets(
	groups []string, measures []api.RequestMeasure,
) (map[string]struct{}, map[string]struct{}) {
	groupSet := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		groupSet[group] = struct{}{}
	}
	measureSet := make(map[string]struct{}, len(measures))
	for _, measure := range measures {
		measureSet[measure.Name()] = struct{}{}
	}
	return groupSet, measureSet
}
