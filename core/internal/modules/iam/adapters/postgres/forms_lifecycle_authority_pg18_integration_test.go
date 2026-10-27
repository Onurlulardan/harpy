//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func applyIAMFormsLifecycleAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	document := iamFoundationDocument{
		path:    iamFormsLifecycleAuthorityMigration + "." + origin + ".sql",
		content: readIAMIdentityMigration(t, origin, iamFormsLifecycleAuthorityMigration),
	}
	applyIAMTenantDocuments(t, ctx, database, iamFoundationTenants[0].schema,
		[]iamFoundationDocument{document})
}

func assertIAMFormsLifecycleAuthority(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(
		t, ctx, database, iamFoundationTenants[0].schema,
	)
	defer connection.Close()
	defer transaction.Rollback()
	rows, err := transaction.QueryContext(ctx, `SELECT resource.resource_key,
pg_catalog.count(*),pg_catalog.bool_and(
 resource.owner_domain='forms' AND resource.resource_kind='system_resource'
 AND resource.is_protected AND resource.state='active'
 AND permission.match_kind='exact' AND NOT permission.is_protected
 AND permission.state='active' AND action.state='active')
FROM iam_resources resource
JOIN iam_permissions permission ON permission.resource_id=resource.id
JOIN iam_actions action ON action.id=permission.action_id
WHERE resource.resource_key IN ('forms.definition','forms.publication')
GROUP BY resource.resource_key ORDER BY resource.resource_key`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	want := map[string]int64{"forms.definition": 4, "forms.publication": 6}
	seen := 0
	for rows.Next() {
		var resource string
		var count int64
		var valid bool
		if err = rows.Scan(&resource, &count, &valid); err != nil {
			t.Fatal(err)
		}
		if count != want[resource] || !valid {
			t.Fatalf("Forms lifecycle authority %s = %d/%t", resource, count, valid)
		}
		seen++
	}
	if err = rows.Err(); err != nil || seen != 2 {
		t.Fatalf("Forms lifecycle authority rows = %d: %v", seen, err)
	}
}
