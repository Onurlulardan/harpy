package api

import (
	"testing"
)

func mustCasefoldTextFieldType(t *testing.T) FieldType {
	t.Helper()
	profile, err := NewTextProfile(
		TextCasefoldUnicodeDefaultV1, TextAccentDisabled,
	)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := NewTextFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func mustFixedFieldType(t *testing.T, kind LogicalType) FieldType {
	t.Helper()
	constructors := map[LogicalType]func() (FieldType, error){
		LogicalTypeBoolean: NewBooleanFieldType, LogicalTypeDate: NewDateFieldType,
		LogicalTypeTimestamp: NewTimestampFieldType, LogicalTypeUUID: NewUUIDFieldType,
		LogicalTypeRelation: NewRelationFieldType, LogicalTypeUUIDList: NewUUIDListFieldType,
		LogicalTypeJSON: NewJSONFieldType,
	}
	fieldType, err := constructors[kind]()
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func TestFieldCapabilitiesCanonicalizeAndOwnCollections(t *testing.T) {
	fieldType := mustCasefoldTextFieldType(t)
	filters := []FilterCapabilityInput{
		{Operator: OperatorIsNull, CountSafe: false},
		{Operator: OperatorEq, CountSafe: true},
		{Operator: OperatorContains, CountSafe: false},
	}
	aggregates := []AggregateFunction{
		AggregateFunctionMax, AggregateFunctionCountDistinct,
	}
	capabilities, err := NewFieldCapabilities(fieldType, FieldCapabilitiesInput{
		Selectable: true, Filters: filters, Sortable: true,
		CursorCanonicalBytes: 256, Searchable: true, Groupable: true,
		Aggregates: aggregates,
	})
	if err != nil || capabilities.Validate() != nil {
		t.Fatalf("valid capabilities rejected: %v", err)
	}
	filters[0].Operator = OperatorContainsAll
	aggregates[0] = AggregateFunctionSum
	actualFilters, _ := capabilities.Filters()
	actualAggregates, _ := capabilities.Aggregates()
	if len(actualFilters) != 3 || actualFilters[0].operator != OperatorEq ||
		actualFilters[1].operator != OperatorContains ||
		actualFilters[2].operator != OperatorIsNull {
		t.Fatal("filter capabilities were not canonical and owned")
	}
	if len(actualAggregates) != 2 ||
		actualAggregates[0] != AggregateFunctionCountDistinct ||
		actualAggregates[1] != AggregateFunctionMax {
		t.Fatal("aggregate capabilities were not canonical and owned")
	}
	actualFilters[0] = FilterCapability{}
	actualAggregates[0] = AggregateFunctionSum
	if _, found, _ := capabilities.Filter(OperatorEq); !found {
		t.Fatal("filter accessor exposed internal storage")
	}
	if allowed, _ := capabilities.AllowsAggregate(AggregateFunctionCountDistinct); !allowed {
		t.Fatal("aggregate accessor exposed internal storage")
	}
	filter, found, err := capabilities.Filter(OperatorEq)
	if err != nil || !found {
		t.Fatal("declared filter was not found")
	}
	operator, operatorErr := filter.Operator()
	countSafe, countErr := filter.CountSafe()
	if operatorErr != nil || countErr != nil || operator != OperatorEq || !countSafe {
		t.Fatal("filter capability did not round trip")
	}
}

func TestFieldCapabilitiesRejectDuplicateOrInvalidDeclarations(t *testing.T) {
	fieldType := mustCasefoldTextFieldType(t)
	inputs := []FieldCapabilitiesInput{
		{Filters: []FilterCapabilityInput{{Operator: OperatorEq}, {Operator: OperatorEq}}},
		{Filters: []FilterCapabilityInput{{Operator: 0}}},
		{Aggregates: []AggregateFunction{AggregateFunctionMax, AggregateFunctionMax}},
		{Aggregates: []AggregateFunction{AggregateFunctionCount}},
		{Sortable: true, CursorCanonicalBytes: 0},
		{Sortable: true, CursorCanonicalBytes: 2049},
		{Sortable: false, CursorCanonicalBytes: 1},
	}
	for _, input := range inputs {
		if _, err := NewFieldCapabilities(fieldType, input); err == nil {
			t.Fatal("invalid field capabilities accepted")
		}
	}
}
