package compiler

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompileKeysetUsesLexicographicNullsLastSemantics(t *testing.T) {
	name, stable, sorts := fixtureKeysetSorts(t)
	nameLast, err := cursor.NewLastValue(
		fixtureScalar(t, api.LogicalTypeText, "m"),
	)
	if err != nil {
		t.Fatal(err)
	}
	stableValue := fixtureScalar(t, api.LogicalTypeUUID,
		"abcdefab-cdef-4abc-8def-abcdefabcdef")
	stableLast, err := cursor.NewLastValue(stableValue)
	if err != nil {
		t.Fatal(err)
	}
	binder := newBinder()
	sql, err := compileKeyset(
		sorts, []cursor.LastValue{nameLast, stableLast}, binder,
	)
	want := `(("fld_name" COLLATE pg_catalog.pg_unicode_fast < $1 OR "fld_name" COLLATE pg_catalog.pg_unicode_fast IS NULL)) OR ("fld_name" COLLATE pg_catalog.pg_unicode_fast = $2 AND ("fld_id" > $3 OR "fld_id" IS NULL))`
	wantArguments := []any{
		"m", "m", "abcdefab-cdef-4abc-8def-abcdefabcdef",
	}
	if err != nil || sql != want || !reflect.DeepEqual(binder.arguments, wantArguments) {
		t.Fatalf("sql=%s args=%#v err=%v", sql, binder.arguments, err)
	}

	binder = newBinder()
	sql, err = compileKeyset(
		[]resolved.Sort{
			fixtureSort(t, name, api.DirectionDescending, false),
			fixtureSort(t, stable, api.DirectionAscending, true),
		},
		[]cursor.LastValue{cursor.NewNullLastValue(), stableLast}, binder,
	)
	want = `("fld_name" COLLATE pg_catalog.pg_unicode_fast IS NULL AND ("fld_id" > $1 OR "fld_id" IS NULL))`
	if err != nil || sql != want || !reflect.DeepEqual(
		binder.arguments, []any{"abcdefab-cdef-4abc-8def-abcdefabcdef"},
	) {
		t.Fatalf("null keyset sql=%s args=%#v err=%v", sql, binder.arguments, err)
	}
	if _, err := compileKeyset(
		sorts, []cursor.LastValue{nameLast, cursor.NewNullLastValue()}, newBinder(),
	); !errors.Is(err, ErrContract) {
		t.Fatal("null stable tie-breaker was accepted")
	}
}

func fixtureKeysetSorts(
	testingT testing.TB,
) (api.FieldDefinition, api.FieldDefinition, []resolved.Sort) {
	testingT.Helper()
	name := fixtureField(testingT,
		"456789ab-4567-4567-8567-456789abcdef", "name", "fld_name",
		fixtureFoldedTextType(testingT, false), "fld_name_folded", "",
		[]api.Operator{}, true, false,
	)
	stable := fixtureField(testingT,
		"56789abc-5678-4678-8678-56789abcdefa", "id", "fld_id",
		fixtureUUIDType(testingT), "", "", []api.Operator{}, true, false,
	)
	sorts := []resolved.Sort{
		fixtureSort(testingT, name, api.DirectionDescending, false),
		fixtureSort(testingT, stable, api.DirectionAscending, true),
	}
	return name, stable, sorts
}
