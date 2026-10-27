package sqlidentifier_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestValidatedTokenPreservesOpaqueAndRenderContracts(t *testing.T) {
	t.Parallel()
	const identifier = "frm_secret"
	token, err := sqlidentifier.MintDynamic(identifier)
	if err != nil {
		t.Fatal(err)
	}
	if err := token.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, output := range []string{
		fmt.Sprint(token), fmt.Sprintf("%+v", token), fmt.Sprintf("%#v", token),
	} {
		if strings.Contains(output, identifier) {
			t.Fatalf("validated token formatting exposed identifier: %q", output)
		}
	}
	if _, err := json.Marshal(token); err == nil || strings.Contains(err.Error(), identifier) {
		t.Fatalf("validated token JSON result was unsafe: %v", err)
	}
	var builder strings.Builder
	if err := sqlidentifier.AppendSQL(token, &builder); err != nil {
		t.Fatal(err)
	}
	if got, want := builder.String(), `"frm_secret"`; got != want {
		t.Fatalf("rendered token = %q, want %q", got, want)
	}
}
