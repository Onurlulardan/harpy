package compiler

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
	"golang.org/x/text/unicode/norm"
)

func FuzzCompilePredicateNeverInterpolatesValues(f *testing.F) {
	for _, seed := range []string{
		`x' OR TRUE --`, `%; DROP TABLE users; --`, `%_\`, "Stra\u00dfe", "\u0130stanbul",
	} {
		f.Add(seed)
	}
	field := fixtureField(f,
		"bcdefabc-bcde-4cde-8cde-bcdefabcdefa", `x"; DROP TABLE fields; --`,
		"fld_text", fixtureFoldedTextType(f, false), "fld_text_folded", "",
		[]api.Operator{api.OperatorEq, api.OperatorContains}, false, false,
	)
	f.Fuzz(func(t *testing.T, raw string) {
		if raw == "" || len(raw) > 4096 || !utf8.ValidString(raw) ||
			strings.ContainsRune(raw, '\x00') || !norm.NFC.IsNormalString(raw) {
			t.Skip()
		}
		exact, err := resolved.NewScalarValue(api.LogicalTypeText, raw)
		if err != nil {
			t.Fatal(err)
		}
		binder := newBinder()
		sql, err := compilePredicate(
			field, api.OperatorEq, []resolved.Value{exact}, binder,
		)
		wantSQL := `"fld_text" COLLATE pg_catalog.pg_unicode_fast = $1`
		if err != nil || sql != wantSQL || !reflect.DeepEqual(binder.arguments, []any{raw}) {
			t.Fatalf("exact SQL=%s args=%#v err=%v", sql, binder.arguments, err)
		}
		canonical := textprofile.UnicodeDefaultCasefoldV1(raw)
		if canonical == "" || strings.ContainsRune(canonical, '\x00') {
			t.Skip()
		}
		pattern, err := resolved.NewScalarValue(api.LogicalTypeText, canonical)
		if err != nil {
			t.Fatal(err)
		}
		binder = newBinder()
		sql, err = compilePredicate(
			field, api.OperatorContains, []resolved.Value{pattern}, binder,
		)
		wantSQL = `"fld_text_folded" COLLATE pg_catalog.pg_unicode_fast LIKE $1 ESCAPE E'\\'`
		wantArguments := []any{"%" + escapeLikeLiteral(canonical) + "%"}
		if err != nil || sql != wantSQL ||
			!reflect.DeepEqual(binder.arguments, wantArguments) {
			t.Fatalf("pattern SQL=%s args=%#v err=%v", sql, binder.arguments, err)
		}
	})
}
