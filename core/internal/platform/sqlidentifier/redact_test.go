package sqlidentifier_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestTokenFormattingIsAlwaysRedacted(t *testing.T) {
	t.Parallel()
	token, err := sqlidentifier.MintDynamic("frm_secret")
	if err != nil {
		t.Fatal(err)
	}
	outputs := []string{
		token.String(), token.GoString(),
		fmt.Sprintf("%s", token), fmt.Sprintf("%q", token),
		fmt.Sprintf("%v", token), fmt.Sprintf("%+v", token),
		fmt.Sprintf("%#v", token), fmt.Sprintf("%x", token),
	}
	for _, output := range outputs {
		if output == "" || strings.Contains(output, "frm_secret") {
			t.Fatalf("token formatting exposed identifier: %q", output)
		}
	}
}
