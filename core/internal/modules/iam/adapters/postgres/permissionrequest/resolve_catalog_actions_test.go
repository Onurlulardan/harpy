package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestResolveCatalogActionsSQLUsesScopedRoutineOnly(t *testing.T) {
	if !strings.Contains(
		permissionRequestCatalogActionsSQL,
		"iam_permission_request_action_catalog_v1(",
	) || !strings.Contains(
		permissionRequestCatalogActionsSQL,
		"(pg_catalog.current_schemas(false))[2]::name",
	) || strings.Contains(permissionRequestCatalogActionsSQL, "FROM iam_actions") {
		t.Fatalf("catalog action SQL bypasses scoped routine: %s", permissionRequestCatalogActionsSQL)
	}
}

func TestResolveCatalogActionsRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewPermissionRequestCatalogStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.ResolveCatalogActions(
		context.Background(), platformtx.TenantScope{},
		permissionrequest.CatalogQuery{
			Kind: permissionrequest.CatalogKindAll, Limit: 10,
		},
	)
	if err == nil {
		t.Fatal("catalog actions accepted an unowned tenant scope")
	}
}
