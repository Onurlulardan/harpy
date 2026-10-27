package api

import "testing"

func TestFieldCapabilityOperatorCeiling(t *testing.T) {
	exactProfile, _ := NewTextProfile(TextCasefoldDisabled, TextAccentDisabled)
	exactText, _ := NewTextFieldType(exactProfile)
	integerProfile, _ := NewIntegerProfile(-10, 10)
	integerType, _ := NewIntegerFieldType(integerProfile)
	textListType, _ := NewTextListFieldType(exactProfile)
	tests := []struct {
		name      string
		fieldType FieldType
		operator  Operator
		allowed   bool
	}{
		{"casefold text contains", mustCasefoldTextFieldType(t), OperatorContains, true},
		{"exact text contains", exactText, OperatorContains, false},
		{"integer between", integerType, OperatorBetween, true},
		{"integer contains", integerType, OperatorContains, false},
		{"boolean order", mustFixedFieldType(t, LogicalTypeBoolean), OperatorGT, false},
		{"uuid in", mustFixedFieldType(t, LogicalTypeUUID), OperatorIn, true},
		{"list contains any", textListType, OperatorContainsAny, true},
		{"list equality", textListType, OperatorEq, false},
		{"json equality", mustFixedFieldType(t, LogicalTypeJSON), OperatorEq, true},
		{"json in", mustFixedFieldType(t, LogicalTypeJSON), OperatorIn, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFieldCapabilities(test.fieldType, FieldCapabilitiesInput{
				Filters: []FilterCapabilityInput{{Operator: test.operator}},
			})
			if (err == nil) != test.allowed {
				t.Fatalf("operator allowed=%v, want %v", err == nil, test.allowed)
			}
		})
	}
}

func TestFieldCapabilityAggregateCeiling(t *testing.T) {
	integerProfile, _ := NewIntegerProfile(-10, 10)
	integerType, _ := NewIntegerFieldType(integerProfile)
	exactProfile, _ := NewTextProfile(TextCasefoldDisabled, TextAccentDisabled)
	textType, _ := NewTextFieldType(exactProfile)
	textListType, _ := NewTextListFieldType(exactProfile)
	tests := []struct {
		name      string
		fieldType FieldType
		function  AggregateFunction
		allowed   bool
	}{
		{"integer distinct", integerType, AggregateFunctionCountDistinct, true},
		{"integer sum", integerType, AggregateFunctionSum, true},
		{"text min", textType, AggregateFunctionMin, true},
		{"text avg", textType, AggregateFunctionAvg, false},
		{"boolean distinct", mustFixedFieldType(t, LogicalTypeBoolean), AggregateFunctionCountDistinct, true},
		{"boolean min", mustFixedFieldType(t, LogicalTypeBoolean), AggregateFunctionMin, false},
		{"json distinct", mustFixedFieldType(t, LogicalTypeJSON), AggregateFunctionCountDistinct, false},
		{"list distinct", textListType, AggregateFunctionCountDistinct, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFieldCapabilities(test.fieldType, FieldCapabilitiesInput{
				Aggregates: []AggregateFunction{test.function},
			})
			if (err == nil) != test.allowed {
				t.Fatalf("aggregate allowed=%v, want %v", err == nil, test.allowed)
			}
		})
	}
}

func TestFieldCapabilitySearchSortAndGroupCeilings(t *testing.T) {
	if _, err := NewFieldCapabilities(mustCasefoldTextFieldType(t), FieldCapabilitiesInput{
		Searchable: true,
	}); err != nil {
		t.Fatalf("casefold text search rejected: %v", err)
	}
	for _, input := range []FieldCapabilitiesInput{
		{Searchable: true},
		{Sortable: true, CursorCanonicalBytes: 32},
		{Groupable: true},
	} {
		if _, err := NewFieldCapabilities(
			mustFixedFieldType(t, LogicalTypeJSON), input,
		); err == nil {
			t.Fatal("JSON exceeded field capability ceiling")
		}
	}
}

func TestOrderedEnumControlsOrderAndMinMaxCapabilities(t *testing.T) {
	ordered := mustEnumFieldType(t, true)
	unordered := mustEnumFieldType(t, false)
	for _, test := range []struct {
		name      string
		fieldType FieldType
		input     FieldCapabilitiesInput
		allowed   bool
	}{
		{"ordered filter", ordered, FieldCapabilitiesInput{Filters: []FilterCapabilityInput{{Operator: OperatorGT}}}, true},
		{"unordered filter", unordered, FieldCapabilitiesInput{Filters: []FilterCapabilityInput{{Operator: OperatorGT}}}, false},
		{"ordered min", ordered, FieldCapabilitiesInput{Aggregates: []AggregateFunction{AggregateFunctionMin}}, true},
		{"unordered min", unordered, FieldCapabilitiesInput{Aggregates: []AggregateFunction{AggregateFunctionMin}}, false},
		{"ordered sort", ordered, FieldCapabilitiesInput{Sortable: true, CursorCanonicalBytes: 32}, true},
		{"unordered sort", unordered, FieldCapabilitiesInput{Sortable: true, CursorCanonicalBytes: 32}, false},
		{"unordered distinct", unordered, FieldCapabilitiesInput{Aggregates: []AggregateFunction{AggregateFunctionCountDistinct}}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFieldCapabilities(test.fieldType, test.input)
			if (err == nil) != test.allowed {
				t.Fatalf("enum capability allowed=%v, want %v", err == nil, test.allowed)
			}
		})
	}
}

func mustEnumFieldType(t *testing.T, ordered bool) FieldType {
	t.Helper()
	var ordinal *int
	if ordered {
		value := 1
		ordinal = &value
	}
	enumValue, err := NewEnumValue("open", ordinal)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := NewEnumProfile(
		"80000000-0000-4000-8000-000000000001", []EnumValue{enumValue},
	)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := NewEnumFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}
