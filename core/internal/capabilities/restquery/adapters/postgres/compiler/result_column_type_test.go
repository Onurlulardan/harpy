package compiler

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestResultColumnPreservesDeclaredDecimalScale(t *testing.T) {
	fieldType := fixtureDecimalType(t, 12, 4)
	field := fixtureField(t,
		"88888888-8888-4888-8888-888888888888", "amount", "fld_amount",
		fieldType, "", "", []api.Operator{api.OperatorEq}, false, false,
	)
	statement := compileProjectionStatement(t, field)
	columns, err := statement.Columns()
	if err != nil || len(columns) != 2 {
		t.Fatalf("columns unavailable: count=%d err=%v", len(columns), err)
	}
	declared, err := columns[0].FieldType()
	if err != nil {
		t.Fatal(err)
	}
	profile, present, err := declared.Decimal()
	if err != nil || !present {
		t.Fatal("decimal profile is missing")
	}
	precision, precisionErr := profile.Precision()
	scale, scaleErr := profile.Scale()
	if precisionErr != nil || scaleErr != nil || precision != 12 || scale != 4 {
		t.Fatal("decimal precision/scale changed")
	}
	binder := newBinder()
	value := fixtureScalar(t, api.LogicalTypeDecimal, "12.3400")
	sql, err := compilePredicate(
		field, api.OperatorEq, []resolved.Value{value}, binder,
	)
	if err != nil || sql != `"fld_amount" = $1` ||
		!reflect.DeepEqual(binder.arguments, []any{"12.3400"}) {
		t.Fatalf("decimal value was coerced: sql=%s args=%#v err=%v", sql, binder.arguments, err)
	}
}

func TestResultColumnPreservesOrderedEnumProfile(t *testing.T) {
	field := fixtureField(t,
		"99999999-9999-4999-8999-999999999999", "status", "fld_status",
		fixtureOrderedEnumType(t), "", "", []api.Operator{}, false, false,
	)
	statement := compileProjectionStatement(t, field)
	columns, err := statement.Columns()
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := columns[0].FieldType()
	if err != nil {
		t.Fatal(err)
	}
	profile, present, err := fieldType.Enum()
	if err != nil || !present {
		t.Fatal("enum profile is missing")
	}
	version, versionErr := profile.Version()
	ordered, orderedErr := profile.Ordered()
	values, valuesErr := profile.Values()
	if versionErr != nil || orderedErr != nil || valuesErr != nil || !ordered ||
		version.String() != "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee" || len(values) != 2 {
		t.Fatal("enum version/order changed")
	}
}

func compileProjectionStatement(
	testingT testing.TB, field api.FieldDefinition,
) RecordStatement {
	testingT.Helper()
	stable := fixtureField(testingT,
		"aaaaaaaa-1111-4111-8111-111111111111", "id", "fld_id",
		fixtureUUIDType(testingT), "", "", []api.Operator{}, true, false,
	)
	query := fixtureRecordQuery(testingT,
		[]resolved.Projection{fixtureProjection(testingT, field, "value", false)}, nil,
		[]resolved.Sort{fixtureSort(
			testingT, stable, api.DirectionAscending, true,
		)}, nil, 1,
	)
	statement, err := compileRecord(fixtureRecordInput(
		testingT, "forms", query, fixtureAllVisibility(testingT),
	))
	if err != nil {
		testingT.Fatal(err)
	}
	return statement
}
