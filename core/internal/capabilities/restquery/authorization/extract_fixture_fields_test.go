package authorization

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func authUUIDField(
	t *testing.T, id, name, columnName string,
) api.FieldDefinition {
	t.Helper()
	fieldType, _ := api.NewUUIDFieldType()
	return authField(t, id, name, fieldType, columnName, nil,
		api.FieldCapabilitiesInput{
			Selectable: true, Sortable: true, CursorCanonicalBytes: 36,
			Filters: []api.FilterCapabilityInput{
				{Operator: api.OperatorEq, CountSafe: true},
			},
			Groupable: true,
			Aggregates: []api.AggregateFunction{
				api.AggregateFunctionCountDistinct,
			},
		})
}

func authTextField(
	t *testing.T, id, name, columnName string,
) api.FieldDefinition {
	t.Helper()
	profile, _ := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	fieldType, _ := api.NewTextFieldType(profile)
	folded, _ := sqlidentifier.MintDynamic("fld_" + columnName + "_folded")
	return authField(t, id, name, fieldType, columnName, &folded,
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

func authDecimalField(
	t *testing.T, id, name, columnName string,
) api.FieldDefinition {
	t.Helper()
	profile, _ := api.NewDecimalProfile(12, 2)
	fieldType, _ := api.NewDecimalFieldType(profile)
	return authField(t, id, name, fieldType, columnName, nil,
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

func authField(
	t *testing.T,
	id string,
	name string,
	fieldType api.FieldType,
	columnName string,
	folded *sqlidentifier.Token,
	input api.FieldCapabilitiesInput,
) api.FieldDefinition {
	t.Helper()
	column, _ := sqlidentifier.MintDynamic("fld_" + columnName)
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
