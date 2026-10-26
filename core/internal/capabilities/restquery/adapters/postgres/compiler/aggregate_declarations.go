package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

func extractAggregateGroups(
	query resolved.AggregateQuery,
) ([]aggregateGroup, error) {
	declarations, err := query.Groups()
	if err != nil {
		return nil, ErrContract
	}
	result := make([]aggregateGroup, len(declarations))
	for index, declaration := range declarations {
		field, fieldErr := declaration.Field()
		key, keyErr := declaration.Key()
		value, valueErr := key.Value()
		if fieldErr != nil || keyErr != nil || valueErr != nil {
			return nil, ErrContract
		}
		result[index] = aggregateGroup{
			declaration: declaration, field: field, key: value,
		}
	}
	return result, nil
}

func extractAggregateMeasures(
	query resolved.AggregateQuery,
) ([]aggregateMeasure, error) {
	declarations, err := query.Measures()
	if err != nil {
		return nil, ErrContract
	}
	result := make([]aggregateMeasure, len(declarations))
	for index, declaration := range declarations {
		field, present, fieldErr := declaration.Field()
		key, keyErr := declaration.Key()
		value, valueErr := key.Value()
		function, functionErr := declaration.Function()
		if fieldErr != nil || keyErr != nil || valueErr != nil || functionErr != nil {
			return nil, ErrContract
		}
		result[index] = aggregateMeasure{
			declaration: declaration, field: field, hasField: present,
			key: value, function: function,
		}
	}
	return result, nil
}
