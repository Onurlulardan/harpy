//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func prepareIAMSystemTransferBase(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	applyIAMFoundation(t, ctx, database, origin)
	seedIAMFoundationBase(t, ctx, database)
	applyIAMLifecycleGuards(t, ctx, database, origin)
	applyIAMRBAC(t, ctx, database, origin)
	permissionRead := iamFoundationDocument{
		path: "20260716013433_create_iam_user_has_permission_v1." + origin + ".sql",
		content: readIAMIdentityMigration(
			t, origin, "20260716013433_create_iam_user_has_permission_v1",
		),
	}
	for _, tenant := range iamFoundationTenants {
		seedIAMRBACCatalog(t, ctx, database, tenant.schema)
		seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
		applyIAMTenantDocuments(t, ctx, database, tenant.schema,
			[]iamFoundationDocument{permissionRead})
	}
}
