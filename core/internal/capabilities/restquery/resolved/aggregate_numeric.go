package resolved

import (
	"fmt"
	"math"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func numericAggregateType(
	fieldType api.FieldType, function api.AggregateFunction,
) (api.FieldType, error) {
	kind, err := fieldType.Kind()
	if err != nil {
		return api.FieldType{}, err
	}
	if kind == api.LogicalTypeInteger {
		if function == api.AggregateFunctionSum {
			profile, profileErr := api.NewIntegerProfile(math.MinInt64, math.MaxInt64)
			if profileErr != nil {
				return api.FieldType{}, profileErr
			}
			return api.NewIntegerFieldType(profile)
		}
		if function == api.AggregateFunctionAvg {
			profile, profileErr := api.NewDecimalProfile(29, 6)
			if profileErr != nil {
				return api.FieldType{}, profileErr
			}
			return api.NewDecimalFieldType(profile)
		}
	}
	if kind != api.LogicalTypeDecimal {
		return api.FieldType{}, fmt.Errorf("resolved REST Query aggregate result is invalid")
	}
	profile, found, err := fieldType.Decimal()
	if err != nil || !found {
		return api.FieldType{}, fmt.Errorf("resolved REST Query aggregate result is invalid")
	}
	precision, precisionErr := profile.Precision()
	scale, scaleErr := profile.Scale()
	if precisionErr != nil || scaleErr != nil {
		return api.FieldType{}, fmt.Errorf("resolved REST Query aggregate result is invalid")
	}
	precision = min(1000, precision+10)
	if function == api.AggregateFunctionAvg {
		scale = min(precision, max(scale, 6))
	}
	resultProfile, profileErr := api.NewDecimalProfile(precision, scale)
	if profileErr != nil {
		return api.FieldType{}, profileErr
	}
	return api.NewDecimalFieldType(resultProfile)
}
