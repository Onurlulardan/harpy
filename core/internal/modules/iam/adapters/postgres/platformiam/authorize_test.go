package platformiampostgres

import (
	"context"
	"strings"
	"testing"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAuthorizeUsesNarrowRoutineAndRejectsUnboundScope(t *testing.T) {
	for _, marker := range []string{
		"iam_platform_authorize_v1", "(pg_catalog.current_schemas(false))[2]::name",
		"$5::text",
	} {
		if !strings.Contains(authorizeSQL, marker) {
			t.Fatalf("authorize SQL lacks %q", marker)
		}
	}
	var storage *store
	if _, err := storage.Authorize(context.Background(), platformtx.TenantScope{},
		platformdomain.EvaluateQuery{}); err == nil {
		t.Fatal("authorize accepted unbound scope")
	}
}
