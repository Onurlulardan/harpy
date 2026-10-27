package compiler

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

const compilerJSONBKey = "de0eddb4-a8e8-4df9-9499-0a581c02f500"

func TestJSONBFieldExpressionUsesOnlySealedStableKey(t *testing.T) {
	field := fixtureJSONBField(
		t, "nickname'); SELECT 1; --", fixtureJSONBExactTextType(t),
		[]api.Operator{api.OperatorEq, api.OperatorIsNull},
	)
	expression, err := baseExpression(field)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(expression, "nickname") ||
		!strings.Contains(expression, `"profile_values" ->> '`+compilerJSONBKey+`'`) {
		t.Fatalf("unsealed JSONB expression %q", expression)
	}
}

func TestJSONBFieldExpressionCastsClosedLogicalTypes(t *testing.T) {
	integer, _ := api.NewIntegerProfile(-20, 20)
	integerType, _ := api.NewIntegerFieldType(integer)
	booleanType, _ := api.NewBooleanFieldType()
	uuidType, _ := api.NewUUIDFieldType()
	jsonType, _ := api.NewJSONFieldType()
	textListType, _ := api.NewTextListFieldType(fixtureJSONBExactTextProfile(t))
	uuidListType, _ := api.NewUUIDListFieldType()
	tests := []struct {
		fieldType api.FieldType
		want      string
	}{
		{integerType, "AS pg_catalog.int8"},
		{booleanType, "AS pg_catalog.boolean"},
		{uuidType, "AS pg_catalog.uuid"},
		{jsonType, `"profile_values" -> '` + compilerJSONBKey + `'`},
		{textListType, "jsonb_array_elements_text"},
		{uuidListType, "CAST(element.value AS pg_catalog.uuid)"},
	}
	for _, test := range tests {
		field := fixtureJSONBField(t, "custom", test.fieldType,
			[]api.Operator{api.OperatorIsNull})
		expression, err := baseExpression(field)
		if err != nil || !strings.Contains(expression, test.want) {
			t.Fatalf("expression %q does not contain %q: %v", expression, test.want, err)
		}
	}
}

func TestJSONBPredicatesPreserveMissingAsNullAndUseContainment(t *testing.T) {
	field := fixtureJSONBField(t, "nickname", fixtureJSONBExactTextType(t), []api.Operator{
		api.OperatorEq, api.OperatorNe, api.OperatorIn,
		api.OperatorNotIn, api.OperatorIsNull, api.OperatorIsNotNull,
	})
	value, _ := resolved.NewScalarValue(api.LogicalTypeText, "Ada")
	tests := []struct {
		op     api.Operator
		values []resolved.Value
		want   string
	}{
		{api.OperatorEq, []resolved.Value{value}, `"profile_values" @> pg_catalog.jsonb_build_object`},
		{api.OperatorNe, []resolved.Value{value}, `"profile_values" ? '` + compilerJSONBKey + `' AND NOT`},
		{api.OperatorIsNull, nil, `NOT ("profile_values" ? '` + compilerJSONBKey + `')`},
		{api.OperatorIsNotNull, nil, `"profile_values" ? '` + compilerJSONBKey + `'`},
	}
	for _, test := range tests {
		statement, err := compilePredicate(field, test.op, test.values, newBinder())
		if err != nil || !strings.Contains(statement, test.want) {
			t.Fatalf("%s statement %q does not contain %q: %v", test.op, statement, test.want, err)
		}
	}
}

func TestJSONBCollectionPredicatesUseIndexableContainment(t *testing.T) {
	typeValue, _ := api.NewUUIDListFieldType()
	field := fixtureJSONBField(t, "external_refs", typeValue, []api.Operator{
		api.OperatorContainsAny, api.OperatorContainsAll,
	})
	first, _ := resolved.NewScalarValue(api.LogicalTypeUUID,
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	second, _ := resolved.NewScalarValue(api.LogicalTypeUUID,
		"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	for _, operator := range []api.Operator{
		api.OperatorContainsAny, api.OperatorContainsAll,
	} {
		statement, err := compilePredicate(
			field, operator, []resolved.Value{first, second}, newBinder(),
		)
		if err != nil || !strings.Contains(statement, " @> ") ||
			strings.Contains(statement, "external_refs") {
			t.Fatalf("unsafe collection predicate %q: %v", statement, err)
		}
	}
}

func fixtureJSONBField(
	t testing.TB,
	publicName string,
	fieldType api.FieldType,
	operators []api.Operator,
) api.FieldDefinition {
	t.Helper()
	document, err := sqlidentifier.MintStatic("profile_values")
	if err != nil {
		t.Fatal(err)
	}
	storage, err := api.NewJSONBFieldStorage(document, compilerJSONBKey)
	if err != nil {
		t.Fatal(err)
	}
	filters := make([]api.FilterCapabilityInput, len(operators))
	for index, operator := range operators {
		filters[index] = api.FilterCapabilityInput{Operator: operator, CountSafe: true}
	}
	capabilities, err := api.NewFieldCapabilities(
		fieldType, api.FieldCapabilitiesInput{
			Selectable: true, Filters: filters, Aggregates: []api.AggregateFunction{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	field, err := api.NewFieldDefinition(
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", publicName,
		fieldType, storage, capabilities,
	)
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func fixtureJSONBExactTextType(t testing.TB) api.FieldType {
	t.Helper()
	fieldType, err := api.NewTextFieldType(fixtureJSONBExactTextProfile(t))
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func fixtureJSONBExactTextProfile(t testing.TB) api.TextProfile {
	t.Helper()
	profile, err := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	if err != nil {
		t.Fatal(err)
	}
	return profile
}
