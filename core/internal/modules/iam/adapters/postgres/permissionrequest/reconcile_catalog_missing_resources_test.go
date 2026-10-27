package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReconcileMissingCatalogResourcesUsesScopedDatabaseDifference(t *testing.T) {
	statement := permissionRequestCatalogMissingResourcesSQL
	if !strings.Contains(statement, "iam_permission_request_resource_catalog_v1(") ||
		!strings.Contains(statement, "(pg_catalog.current_schemas(false))[2]::name") ||
		strings.Contains(statement, "FROM iam_permissions") {
		t.Fatalf("missing catalog SQL bypasses scoped routine: %s", statement)
	}
}

func TestReconcileMissingCatalogResourcesRejectsUnownedTenantScope(t *testing.T) {
	store, err := NewPermissionRequestCatalogStore()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.ReconcileMissingCatalogResources(
		context.Background(), platformtx.TenantScope{},
		permissionrequest.MissingCatalogQuery{
			CatalogQuery: permissionrequest.CatalogQuery{
				Kind: permissionrequest.CatalogKindForm, Limit: 10,
			},
			ActionKey: "read",
		},
	)
	if err == nil {
		t.Fatal("missing catalog accepted an unowned tenant scope")
	}
}
