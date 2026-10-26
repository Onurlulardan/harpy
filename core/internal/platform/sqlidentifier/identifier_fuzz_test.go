package sqlidentifier_test

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func FuzzMintedIdentifierRendersAsOneQuotedToken(f *testing.F) {
	for _, seed := range []string{
		"forms_definitions", "frm_orders", `name"; DROP TABLE users; --`, "", "é",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		minters := []func(string) (sqlidentifier.Token, error){
			sqlidentifier.MintDynamic,
			sqlidentifier.MintStatic,
		}
		for _, mint := range minters {
			token, err := mint(value)
			if err != nil {
				continue
			}
			var builder strings.Builder
			if err := sqlidentifier.AppendSQL(token, &builder); err != nil {
				t.Fatal(err)
			}
			rendered := builder.String()
			if len(rendered) < 3 || rendered[0] != '"' ||
				rendered[len(rendered)-1] != '"' || strings.Count(rendered, `"`) != 2 {
				t.Fatalf("unsafe identifier rendering: %q", rendered)
			}
		}
	})
}
