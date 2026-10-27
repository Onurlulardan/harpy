package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetCatalogResourcesSQLUsesScopedRoutineOnly(t *testing.T) {
	statement := permissionRequestCatalogResourcesSQL
	if !strings.Contains(statement, "iam_permission_request_resource_catalog_v1(") ||
		!strings.Contains(statement, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatalf("catalog resource SQL bypasses scoped routine: %s", statement)
	}
	for _, direct := range []string{"FROM iam_resources", "FROM iam_permissions"} {
		if strings.Contains(statement, direct) {
			t.Fatalf("catalog resource SQL contains direct table read %q", direct)
		}
	}
}

func TestGetCatalogResourcesRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewPermissionRequestCatalogStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.GetCatalogResources(
		context.Background(), platformtx.TenantScope{},
		permissionrequest.CatalogQuery{
			Kind: permissionrequest.CatalogKindAll, Limit: 10,
		},
	)
	if err == nil {
		t.Fatal("catalog resources accepted an unowned tenant scope")
	}
}
