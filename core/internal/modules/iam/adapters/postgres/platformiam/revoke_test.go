package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRevokeUsesNarrowRoutineAndRejectsUnboundScope(t *testing.T) {
	for _, marker := range []string{
		"iam_platform_operator_revoke_v1", "(pg_catalog.current_schemas(false))[2]::name",
		"$15::text", "$9::text", "$12::boolean",
	} {
		if !strings.Contains(revokeSQL, marker) {
			t.Fatalf("revoke SQL lacks %q", marker)
		}
	}
	var storage *store
	if _, err := storage.Revoke(context.Background(), platformtx.TenantScope{},
		platformdomain.OperatorRevoke{}); err == nil {
		t.Fatal("revoke accepted unbound scope")
	}
}
