//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func applyIAMRBACRuntime(t *testing.T, ctx context.Context, database *sql.DB, origin string) {
	t.Helper()
	names := append([]string{"20260716013433_create_iam_user_has_permission_v1"},
		iamRBACRuntimeMigrationNames...)
	documents := make([]iamFoundationDocument, 0, len(names))
	for _, name := range names {
		documents = append(documents, iamFoundationDocument{
			path: name + "." + origin + ".sql", content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}
