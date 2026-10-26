package catalogowner

import (
	"fmt"
	"strings"

	nativequery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/nativequery"
)

func nativeJoinAggregation(
	definition endpointDefinition, fields []nativequery.Field,
) (*nativequery.Aggregation, error) {
	if !definition.Aggregation.Enabled {
		return nil, nil
	}
	result := &nativequery.Aggregation{
		Groups: make([]nativequery.Field, len(definition.Aggregation.GroupBy)),
		Measures: make([]nativequery.Measure,
			len(definition.Aggregation.Aggregates)),
	}
	for index, reference := range definition.Aggregation.GroupBy {
		field, ok := nativeJoinFieldReference(reference, fields)
		if !ok {
			return nil, fmt.Errorf("Analytics native group reference is invalid")
		}
		result.Groups[index] = field
	}
	for index, aggregate := range definition.Aggregation.Aggregates {
		function := strings.ToLower(aggregate.Function)
		measure := nativequery.Measure{Alias: aggregate.Alias, Function: function}
		if function != "count" {
			field, ok := nativeJoinFieldReference(aggregate.Field, fields)
			if !ok {
				return nil, fmt.Errorf("Analytics native measure reference is invalid")
			}
			measure.Source, measure.Field = field.Source, field.Field
		}
		result.Measures[index] = measure
	}
	return result, nil
}

func nativeJoinOrders(
	values []endpointOrder, fields []nativequery.Field,
	aggregation *nativequery.Aggregation,
) ([]nativequery.Order, error) {
	result := make([]nativequery.Order, len(values))
	measures := map[string]bool{}
	if aggregation != nil {
		for _, value := range aggregation.Measures {
			measures[value.Alias] = true
		}
	}
	for index, value := range values {
		direction := strings.ToLower(value.Direction)
		if measures[value.Field] {
			result[index] = nativequery.Order{
				Measure: value.Field, Direction: direction,
			}
			continue
		}
		field, ok := nativeJoinFieldReference(value.Field, fields)
		if !ok {
			return nil, fmt.Errorf("Analytics native order reference is invalid")
		}
		result[index] = nativequery.Order{
			Source: field.Source, Field: field.Field, Direction: direction,
		}
	}
	return result, nil
}

func nativeJoinFieldReference(
	reference string, fields []nativequery.Field,
) (nativequery.Field, bool) {
	var matched nativequery.Field
	for _, field := range fields {
		if reference == field.Source+"."+field.Field ||
			reference == field.Alias {
			if matched.Alias != "" {
				return nativequery.Field{}, false
			}
			matched = field
		}
	}
	return matched, matched.Alias != ""
}
