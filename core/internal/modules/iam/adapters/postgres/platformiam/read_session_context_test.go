package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReadSessionContextUsesNarrowRoutine(t *testing.T) {
	for _, marker := range []string{
		"iam_platform_session_context_v1", "(pg_catalog.current_schemas(false))[2]::name",
		"$3::bigint",
	} {
		if !strings.Contains(readSessionContextSQL, marker) {
			t.Fatalf("session context SQL lacks %q", marker)
		}
	}
	var storage *store
	if _, err := storage.ReadSessionContext(context.Background(),
		platformtx.TenantScope{}, platformdomain.SessionContextQuery{}); err == nil {
		t.Fatal("session context accepted unbound scope")
	}
}
