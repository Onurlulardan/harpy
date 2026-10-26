package sqlidentifier_test

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestMintDynamicAcceptsOnlyReservedNamespaces(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"frm_orders", "fld_total_g1", "rel_customer"} {
		if _, err := sqlidentifier.MintDynamic(value); err != nil {
			t.Fatalf("valid dynamic identifier %q rejected: %v", value, err)
		}
	}
	for _, value := range []string{
		"", "forms_orders", "frm-quote", "FRM_ORDERS", "frm_ç", "1frm_orders",
	} {
		if _, err := sqlidentifier.MintDynamic(value); err == nil {
			t.Fatalf("invalid dynamic identifier %q accepted", value)
		}
	}
}
