package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func aggregateRawMeasure(
	measure aggregateMeasure, binder *binder,
) (string, error) {
	if binder == nil || measure.declaration.Validate() != nil {
		return "", ErrContract
	}
	switch measure.function {
	case api.AggregateFunctionCount:
		return "pg_catalog.count(*)", nil
	case api.AggregateFunctionCountDistinct:
		expression, err := exactExpression(measure.field)
		if err != nil {
			return "", err
		}
		return "pg_catalog.count(DISTINCT " + expression + ")", nil
	case api.AggregateFunctionSum, api.AggregateFunctionAvg:
		expression, err := baseExpression(measure.field)
		if err != nil {
			return "", err
		}
		return "pg_catalog.sum(CAST(" + expression + " AS pg_catalog.numeric))", nil
	case api.AggregateFunctionMin, api.AggregateFunctionMax:
		return aggregateRawExtreme(measure, binder)
	default:
		return "", ErrContract
	}
}

func aggregateAverageCount(measure aggregateMeasure) (string, error) {
	if measure.function != api.AggregateFunctionAvg || !measure.hasField {
		return "", ErrContract
	}
	expression, err := baseExpression(measure.field)
	if err != nil {
		return "", err
	}
	return "pg_catalog.count(" + expression + ")", nil
}

func aggregateRawExtreme(
	measure aggregateMeasure, binder *binder,
) (string, error) {
	fieldType, err := measure.field.Type()
	if err != nil {
		return "", ErrContract
	}
	kind, err := fieldType.Kind()
	if err != nil {
		return "", ErrContract
	}
	if kind == api.LogicalTypeEnum {
		return aggregateEnumExtreme(measure, binder)
	}
	expression, err := exactExpression(measure.field)
	if err != nil {
		return "", err
	}
	function := "pg_catalog.min"
	if measure.function == api.AggregateFunctionMax {
		function = "pg_catalog.max"
	}
	return function + "(" + expression + ")", nil
}

func aggregateEnumExtreme(
	measure aggregateMeasure, binder *binder,
) (string, error) {
	values, err := orderedEnumValues(measure.field)
	if err != nil {
		return "", err
	}
	codes := make([]string, len(values))
	for index, value := range values {
		codes[index] = value.code
	}
	placeholder, err := binder.bind(codes)
	if err != nil {
		return "", err
	}
	column, err := baseExpression(measure.field)
	if err != nil {
		return "", err
	}
	function := "pg_catalog.min"
	if measure.function == api.AggregateFunctionMax {
		function = "pg_catalog.max"
	}
	array := "CAST(" + placeholder + " AS pg_catalog.text[])"
	position := "pg_catalog.array_position(" + array + ", " + column + ")"
	return "(" + array + ")[" + function + "(" + position + ")]", nil
}
