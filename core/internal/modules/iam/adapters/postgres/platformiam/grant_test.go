package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGrantUsesNarrowRoutineAndRejectsUnboundScope(t *testing.T) {
	for _, marker := range []string{
		"iam_platform_operator_grant_v1", "(pg_catalog.current_schemas(false))[2]::name",
		"$19::text", "$9::text", "$12::boolean",
	} {
		if !strings.Contains(grantSQL, marker) {
			t.Fatalf("grant SQL lacks %q", marker)
		}
	}
	var storage *store
	if _, err := storage.Grant(context.Background(), platformtx.TenantScope{},
		platformdomain.OperatorGrant{}); err == nil {
		t.Fatal("grant accepted unbound scope")
	}
}
