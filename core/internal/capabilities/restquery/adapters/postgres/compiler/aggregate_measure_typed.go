package compiler

import (
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func aggregateTypedMeasure(
	measure aggregateMeasure, ordinal int, binder *binder,
) (string, error) {
	alias, err := aggregateMeasureAlias(ordinal)
	if err != nil {
		return "", err
	}
	resultType, err := measure.declaration.ResultType()
	if err != nil {
		return "", ErrContract
	}
	kind, err := resultType.Kind()
	if err != nil {
		return "", ErrContract
	}
	if measure.function == api.AggregateFunctionSum && kind == api.LogicalTypeInteger {
		return "CAST(" + alias + " AS pg_catalog.int8)", nil
	}
	if kind != api.LogicalTypeDecimal {
		return alias, nil
	}
	profile, found, err := resultType.Decimal()
	if err != nil || !found {
		return "", ErrContract
	}
	precision, precisionErr := profile.Precision()
	scale, scaleErr := profile.Scale()
	if precisionErr != nil || scaleErr != nil {
		return "", ErrContract
	}
	typed := alias
	if measure.function == api.AggregateFunctionSum ||
		measure.function == api.AggregateFunctionAvg {
		divisor := "CAST(1 AS pg_catalog.numeric)"
		if measure.function == api.AggregateFunctionAvg {
			countAlias, countErr := aggregateAverageCountAlias(ordinal)
			if countErr != nil {
				return "", countErr
			}
			divisor = "CAST(" + countAlias + " AS pg_catalog.numeric)"
		}
		typed, err = aggregateRoundEven(alias, divisor, ordinal, scale, binder)
		if err != nil {
			return "", err
		}
	}
	return "CAST(" + typed + " AS pg_catalog.numeric(" +
		strconv.Itoa(precision) + "," + strconv.Itoa(scale) + "))", nil
}
