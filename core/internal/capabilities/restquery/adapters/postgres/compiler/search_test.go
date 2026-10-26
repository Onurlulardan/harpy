package compiler

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestCompileSearchUsesDeclaredCompanionAndOR(t *testing.T) {
	folded := fixtureField(t,
		"9abcdefa-9abc-4abc-8abc-9abcdefabcde", "name", "fld_name",
		fixtureFoldedTextType(t, false), "fld_name_folded", "",
		[]api.Operator{api.OperatorContains}, false, true,
	)
	accent := fixtureField(t,
		"abcdefab-abcd-4bcd-8bcd-abcdefabcdef", "city", "fld_city",
		fixtureFoldedTextType(t, true), "fld_city_folded", "fld_city_accent",
		[]api.Operator{api.OperatorContains}, false, true,
	)
	terms := []resolved.SearchTerm{
		fixtureSearchTerm(t, folded, "needle"),
		fixtureSearchTerm(t, accent, "cafe"),
	}
	binder := newBinder()
	sql, err := compileSearch(terms, binder)
	want := `("fld_name_folded" COLLATE pg_catalog.pg_unicode_fast LIKE $1 ESCAPE E'\\') OR ("fld_city_accent" COLLATE pg_catalog.pg_unicode_fast LIKE $2 ESCAPE E'\\')`
	if err != nil || sql != want ||
		!reflect.DeepEqual(binder.arguments, []any{"%needle%", "%cafe%"}) {
		t.Fatalf("search sql=%s args=%#v err=%v", sql, binder.arguments, err)
	}
}

func fixtureSearchTerm(
	testingT testing.TB, field api.FieldDefinition, value string,
) resolved.SearchTerm {
	testingT.Helper()
	term, err := resolved.NewSearchTerm(
		field, fixtureScalar(testingT, api.LogicalTypeText, value),
		fixturePath(testingT, "/search"),
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return term
}
