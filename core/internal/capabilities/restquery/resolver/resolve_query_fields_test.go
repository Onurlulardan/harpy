package resolver

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func fixtureUUIDField(
	t *testing.T, id, name, columnName string,
) api.FieldDefinition {
	t.Helper()
	fieldType, _ := api.NewUUIDFieldType()
	return fixtureField(t, id, name, fieldType, columnName, nil,
		api.FieldCapabilitiesInput{
			Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
			Filters: []api.FilterCapabilityInput{
				{Operator: api.OperatorEq, CountSafe: true},
			},
			Groupable:  true,
			Aggregates: []api.AggregateFunction{api.AggregateFunctionCountDistinct},
		})
}

func fixtureTextField(
	t *testing.T, id, name, suffix string,
) api.FieldDefinition {
	t.Helper()
	profile, err := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, _ := api.NewTextFieldType(profile)
	folded, _ := sqlidentifier.MintDynamic("fld_" + suffix + "_folded")
	return fixtureField(t, id, name, fieldType, "fld_"+suffix, &folded,
		api.FieldCapabilitiesInput{
			Selectable: true, Sortable: true, CursorCanonicalBytes: 256,
			Searchable: true, Groupable: true,
			Filters: []api.FilterCapabilityInput{
				{Operator: api.OperatorEq, CountSafe: true},
				{Operator: api.OperatorContains, CountSafe: true},
			},
			Aggregates: []api.AggregateFunction{
				api.AggregateFunctionCountDistinct,
				api.AggregateFunctionMin,
				api.AggregateFunctionMax,
			},
		})
}

func fixtureDecimalField(
	t *testing.T, id, name, columnName string,
) api.FieldDefinition {
	t.Helper()
	profile, err := api.NewDecimalProfile(12, 2)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, _ := api.NewDecimalFieldType(profile)
	return fixtureField(t, id, name, fieldType, columnName, nil,
		api.FieldCapabilitiesInput{
			Selectable: true, Sortable: true, CursorCanonicalBytes: 64,
			Groupable: true,
			Filters: []api.FilterCapabilityInput{
				{Operator: api.OperatorEq, CountSafe: true},
				{Operator: api.OperatorGT, CountSafe: true},
			},
			Aggregates: []api.AggregateFunction{
				api.AggregateFunctionCountDistinct,
				api.AggregateFunctionSum,
				api.AggregateFunctionAvg,
				api.AggregateFunctionMin,
				api.AggregateFunctionMax,
			},
		})
}

func fixtureField(
	t *testing.T,
	id string,
	name string,
	fieldType api.FieldType,
	columnName string,
	folded *sqlidentifier.Token,
	input api.FieldCapabilitiesInput,
) api.FieldDefinition {
	t.Helper()
	column, _ := sqlidentifier.MintDynamic(columnName)
	storage, err := api.NewFieldStorage(column, folded, nil)
	if err != nil {
		t.Fatal(err)
	}
	capabilities, err := api.NewFieldCapabilities(fieldType, input)
	if err != nil {
		t.Fatal(err)
	}
	field, err := api.NewFieldDefinition(id, name, fieldType, storage, capabilities)
	if err != nil {
		t.Fatal(err)
	}
	return field
}
