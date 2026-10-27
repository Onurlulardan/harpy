package compiler

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func FuzzCompileExactCountNeverInterpolatesValues(f *testing.F) {
	for _, seed := range []string{
		`x' OR TRUE --`, `%; DROP TABLE records; --`, `" UNION SELECT`, "Stra\u00dfe",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		if raw == "" || len(raw) > 4096 || !utf8.ValidString(raw) ||
			strings.ContainsRune(raw, '\x00') || !norm.NFC.IsNormalString(raw) {
			t.Skip()
		}
		statement, err := compileExactCount(fixtureExactQueryWithValue(t, raw))
		if err != nil {
			t.Fatal(err)
		}
		sql, sqlErr := statement.SQL()
		arguments, argumentErr := statement.Arguments()
		wantSQL := `SELECT COUNT(*) FROM "frm_records" WHERE ("fld_text" COLLATE pg_catalog.pg_unicode_fast = $1)`
		if sqlErr != nil || argumentErr != nil || sql != wantSQL ||
			!reflect.DeepEqual(arguments, []any{raw}) {
			t.Fatalf("exact count SQL=%s args=%#v", sql, arguments)
		}
	})
}
