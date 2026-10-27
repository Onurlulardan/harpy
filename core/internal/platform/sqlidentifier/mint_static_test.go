package sqlidentifier_test

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestMintStaticRejectsDynamicNamespaces(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"forms_definitions", "iam_users", "platform_jobs"} {
		if _, err := sqlidentifier.MintStatic(value); err != nil {
			t.Fatalf("valid static identifier %q rejected: %v", value, err)
		}
	}
	for _, value := range []string{
		"", "frm_orders", "fld_total", "rel_customer", "Forms_users",
		"forms-users", "forms_users;drop", strings.Repeat("a", 64),
	} {
		if _, err := sqlidentifier.MintStatic(value); err == nil {
			t.Fatalf("invalid static identifier %q accepted", value)
		}
	}
}
