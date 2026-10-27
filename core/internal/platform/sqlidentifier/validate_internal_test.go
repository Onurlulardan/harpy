package sqlidentifier

import (
	"strings"
	"testing"
)

func TestValidateRejectsPackageInternalTokenTampering(t *testing.T) {
	t.Parallel()
	tests := []Token{
		{identifier: "frm_secret", kind: tokenKindStatic},
		{identifier: "forms_secret", kind: tokenKindDynamic},
		{identifier: "forms-secret", kind: tokenKindStatic},
		{identifier: "forms_secret", kind: tokenKind(99)},
		{identifier: "must_not_leak"},
	}
	for _, token := range tests {
		err := token.Validate()
		if err == nil {
			t.Fatal("tampered SQL identifier token validated")
		}
		if strings.Contains(err.Error(), token.identifier) {
			t.Fatalf("validation error exposed identifier: %v", err)
		}
	}
}

func TestValidateRejectsMutationAfterMint(t *testing.T) {
	t.Parallel()
	token, err := MintDynamic("frm_secret")
	if err != nil {
		t.Fatal(err)
	}
	token.identifier = "forms_secret"
	if err := token.Validate(); err == nil || strings.Contains(err.Error(), token.identifier) {
		t.Fatalf("mutated token returned unsafe result: %v", err)
	}
}
