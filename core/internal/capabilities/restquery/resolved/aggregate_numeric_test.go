package resolved

import (
	"math"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestNumericAggregateTypeRulesAreExact(t *testing.T) {
	integerProfile, _ := api.NewIntegerProfile(-10, 10)
	integerType, _ := api.NewIntegerFieldType(integerProfile)
	sumInteger, err := numericAggregateType(integerType, api.AggregateFunctionSum)
	if err != nil {
		t.Fatal(err)
	}
	sumProfile, found, _ := sumInteger.Integer()
	minimum, _ := sumProfile.Minimum()
	maximum, _ := sumProfile.Maximum()
	if !found || minimum != math.MinInt64 || maximum != math.MaxInt64 {
		t.Fatal("sum(integer) result is not signed-64")
	}
	averageInteger, err := numericAggregateType(integerType, api.AggregateFunctionAvg)
	if err != nil {
		t.Fatal(err)
	}
	assertDecimalProfile(t, averageInteger, 29, 6)
	decimalProfile, _ := api.NewDecimalProfile(995, 990)
	decimalType, _ := api.NewDecimalFieldType(decimalProfile)
	sumDecimal, _ := numericAggregateType(decimalType, api.AggregateFunctionSum)
	averageDecimal, _ := numericAggregateType(decimalType, api.AggregateFunctionAvg)
	assertDecimalProfile(t, sumDecimal, 1000, 990)
	assertDecimalProfile(t, averageDecimal, 1000, 990)
}

func assertDecimalProfile(
	t *testing.T, fieldType api.FieldType, wantPrecision, wantScale int,
) {
	t.Helper()
	profile, found, err := fieldType.Decimal()
	if err != nil || !found {
		t.Fatalf("decimal result missing: %v", err)
	}
	precision, _ := profile.Precision()
	scale, _ := profile.Scale()
	if precision != wantPrecision || scale != wantScale {
		t.Fatalf("decimal(%d,%d), want decimal(%d,%d)",
			precision, scale, wantPrecision, wantScale)
	}
}
