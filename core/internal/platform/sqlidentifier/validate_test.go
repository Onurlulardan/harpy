package sqlidentifier_test

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestMintedStaticAndDynamicTokensValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
		mint  func(string) (sqlidentifier.Token, error)
	}{
		{"static", "forms_definitions", sqlidentifier.MintStatic},
		{"dynamic table", "frm_orders", sqlidentifier.MintDynamic},
		{"dynamic field", "fld_total_g1", sqlidentifier.MintDynamic},
		{"dynamic relation", "rel_customer", sqlidentifier.MintDynamic},
	}
	for _, test := range tests {
		token, err := test.mint(test.value)
		if err != nil {
			t.Fatalf("%s mint failed: %v", test.name, err)
		}
		if err := token.Validate(); err != nil {
			t.Fatalf("%s token validation failed: %v", test.name, err)
		}
	}
}

func TestZeroTokenValidationFailsClosed(t *testing.T) {
	t.Parallel()
	var zero sqlidentifier.Token
	if err := zero.Validate(); err == nil {
		t.Fatal("zero SQL identifier token validated")
	}
}
