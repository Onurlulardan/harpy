//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"slices"
	"testing"
)

func assertIAMRBACCatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	want := []string{
		"iam_actions",
		"iam_authorization_clock",
		"iam_authorization_operation_receipts",
		"iam_organization_permission_bindings",
		"iam_permission_implications",
		"iam_permissions",
		"iam_protected_admin_policies",
		"iam_resources",
		"iam_role_permission_bindings",
		"iam_user_permission_bindings",
	}
	for _, tenant := range iamFoundationTenants {
		rows, err := database.QueryContext(ctx, `
SELECT tablename
FROM pg_catalog.pg_tables
WHERE schemaname = $1 AND tablename = ANY($2)
ORDER BY tablename`, tenant.schema, want)
		if err != nil {
			t.Fatalf("read %s IAM RBAC catalog: %v", tenant.schema, err)
		}
		var got []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				_ = rows.Close()
				t.Fatalf("scan %s IAM RBAC catalog: %v", tenant.schema, err)
			}
			got = append(got, name)
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("close %s IAM RBAC catalog: %v", tenant.schema, err)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("%s IAM RBAC tables = %v, want %v", tenant.schema, got, want)
		}
	}
}
