package catalogowner

import (
	"fmt"
	"strings"
)

func endpointAggregation(
	definition endpointDefinition, alias string,
) (map[string]any, error) {
	if len(definition.Aggregation.Aggregates) == 0 ||
		len(definition.Aggregation.Aggregates) > 8 ||
		len(definition.Aggregation.GroupBy) > 3 ||
		len(definition.Aggregation.OrderBy) > 3 {
		return nil, fmt.Errorf("endpoint aggregation exceeds the V2 contract")
	}
	groups := make([]string, len(definition.Aggregation.GroupBy))
	groupSet := make(map[string]bool, len(groups))
	for index, value := range definition.Aggregation.GroupBy {
		field, ok := unqualified(value, alias)
		if !ok || groupSet[field] {
			return nil, fmt.Errorf("endpoint aggregation group is invalid")
		}
		groups[index], groupSet[field] = field, true
	}
	measures := make([]map[string]any, len(definition.Aggregation.Aggregates))
	measureSet := make(map[string]bool, len(measures))
	for index, value := range definition.Aggregation.Aggregates {
		function := strings.ToLower(value.Function)
		if value.Alias == "" || measureSet[value.Alias] ||
			(function != "count" && function != "sum" &&
				function != "avg" && function != "min" && function != "max") {
			return nil, fmt.Errorf("endpoint aggregate is invalid")
		}
		measure := map[string]any{"name": value.Alias, "function": function}
		if function != "count" {
			field, ok := unqualified(value.Field, alias)
			if !ok {
				return nil, fmt.Errorf("endpoint aggregate field is invalid")
			}
			measure["field"] = field
		}
		measures[index], measureSet[value.Alias] = measure, true
	}
	sorts := make([]map[string]any, len(definition.Aggregation.OrderBy))
	for index, value := range definition.Aggregation.OrderBy {
		direction := strings.ToLower(value.Direction)
		if direction != "asc" && direction != "desc" {
			return nil, fmt.Errorf("endpoint aggregate sort is invalid")
		}
		sort := map[string]any{"direction": direction}
		if measureSet[value.Field] {
			sort["measure"] = value.Field
		} else {
			field, ok := unqualified(value.Field, alias)
			if !ok || !groupSet[field] {
				return nil, fmt.Errorf("endpoint aggregate sort reference is invalid")
			}
			sort["group"] = field
		}
		sorts[index] = sort
	}
	result := map[string]any{"measures": measures, "limit": 100}
	if len(groups) > 0 {
		result["group_by"] = groups
	}
	if len(sorts) > 0 {
		result["sort"] = sorts
	}
	return result, nil
}
