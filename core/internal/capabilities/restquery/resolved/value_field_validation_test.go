package resolved

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestValueConformanceEnforcesNumericProfiles(t *testing.T) {
	integerProfile, _ := api.NewIntegerProfile(-2, 2)
	integerType, _ := api.NewIntegerFieldType(integerProfile)
	inRange, _ := NewScalarValue(api.LogicalTypeInteger, "2")
	outOfRange, _ := NewScalarValue(api.LogicalTypeInteger, "3")
	if !valueConformsToFieldType(integerType, api.OperatorEq, inRange) ||
		valueConformsToFieldType(integerType, api.OperatorEq, outOfRange) {
		t.Fatal("integer profile bounds were not enforced")
	}
	decimalProfile, _ := api.NewDecimalProfile(4, 2)
	decimalType, _ := api.NewDecimalFieldType(decimalProfile)
	canonical, _ := NewScalarValue(api.LogicalTypeDecimal, "12.30")
	wrongScale, _ := NewScalarValue(api.LogicalTypeDecimal, "12.3")
	overflow, _ := NewScalarValue(api.LogicalTypeDecimal, "123.45")
	if !valueConformsToFieldType(decimalType, api.OperatorEq, canonical) ||
		valueConformsToFieldType(decimalType, api.OperatorEq, wrongScale) ||
		valueConformsToFieldType(decimalType, api.OperatorEq, overflow) {
		t.Fatal("decimal precision or canonical scale was not enforced")
	}
}

func TestValueConformanceEnforcesEnumAndPatternProfiles(t *testing.T) {
	first, _ := api.NewEnumValue("OPEN", nil)
	profile, _ := api.NewEnumProfile(
		"80000000-0000-4000-8000-000000000001", []api.EnumValue{first},
	)
	enumType, _ := api.NewEnumFieldType(profile)
	allowed, _ := NewScalarValue(api.LogicalTypeEnum, "OPEN")
	unknown, _ := NewScalarValue(api.LogicalTypeEnum, "CLOSED")
	if !valueConformsToFieldType(enumType, api.OperatorEq, allowed) ||
		valueConformsToFieldType(enumType, api.OperatorEq, unknown) {
		t.Fatal("enum membership was not enforced")
	}
	textProfile, _ := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	textType, _ := api.NewTextFieldType(textProfile)
	folded, _ := NewScalarValue(api.LogicalTypeText, "ada")
	unfolded, _ := NewScalarValue(api.LogicalTypeText, "Ada")
	if !valueConformsToFieldType(textType, api.OperatorContains, folded) ||
		valueConformsToFieldType(textType, api.OperatorContains, unfolded) {
		t.Fatal("pattern casefold normalization was not enforced")
	}
}
