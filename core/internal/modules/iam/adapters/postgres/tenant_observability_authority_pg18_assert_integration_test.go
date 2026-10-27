//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMTenantObservabilityAuthority(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		connection, transaction := beginIAMTenantTx(
			t, ctx, database, tenant.schema,
		)
		assertIAMTenantObservabilityCatalog(t, ctx, transaction)
		assertIAMTenantObservabilityGrants(t, ctx, transaction)
		assertIAMTenantObservabilityClock(t, ctx, transaction)
		_ = transaction.Rollback()
		_ = connection.Close()
	}
}

func assertIAMTenantObservabilityCatalog(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
) {
	t.Helper()
	var changelogResource, usageResource, readAction string
	var changelogPermission, usagePermission string
	err := transaction.QueryRowContext(ctx, `
SELECT
 (SELECT id::text FROM iam_resources WHERE resource_key='changelog'
  AND owner_domain='audit' AND resource_kind='audit_changelog'
  AND is_protected AND state='active'),
 (SELECT id::text FROM iam_resources WHERE resource_key='analytics.usage'
  AND owner_domain='usage' AND resource_kind='usage_analytics'
  AND is_protected AND state='active'),
 (SELECT id::text FROM iam_actions WHERE action_key='read'
  AND NOT is_protected AND state='active'),
 (SELECT permission.id::text FROM iam_permissions AS permission
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  JOIN iam_actions AS action ON action.id=permission.action_id
  WHERE resource.resource_key='changelog' AND action.action_key='read'
    AND permission.match_kind='exact' AND permission.is_protected
    AND permission.state='active'),
 (SELECT permission.id::text FROM iam_permissions AS permission
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  JOIN iam_actions AS action ON action.id=permission.action_id
  WHERE resource.resource_key='analytics.usage' AND action.action_key='read'
    AND permission.match_kind='exact' AND permission.is_protected
    AND permission.state='active')`).Scan(
		&changelogResource, &usageResource, &readAction,
		&changelogPermission, &usagePermission,
	)
	if err != nil || changelogResource != "019fb202-0000-8000-8000-000000000001" ||
		usageResource != "019fb202-0000-8000-8000-000000000002" ||
		readAction != iamRBACReadActionID ||
		changelogPermission != "019fb202-0000-8000-8000-000000000201" ||
		usagePermission != "019fb202-0000-8000-8000-000000000202" {
		t.Fatalf("tenant observability catalog=%s/%s/%s/%s/%s: %v",
			changelogResource, usageResource, readAction,
			changelogPermission, usagePermission, err)
	}
}

func assertIAMTenantObservabilityGrants(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
) {
	t.Helper()
	var changelog, usage, mismatches, unrelated, protected int
	err := transaction.QueryRowContext(ctx, `
WITH expected(role_name,changelog_count,usage_count) AS (VALUES
 ('Temel Kullanıcı',1,0),('Tenant Yöneticisi',1,1),('Genel Müdür',1,0),
 ('Kullanıcı',1,0),('Yönetici',1,1),('Müdür',1,0),('BT Uzmanı',0,0)
), actual AS (
 SELECT role.name,
  pg_catalog.count(*) FILTER (WHERE resource.resource_key='changelog') AS changelog_count,
  pg_catalog.count(*) FILTER (WHERE resource.resource_key='analytics.usage') AS usage_count
 FROM iam_roles AS role
 LEFT JOIN iam_role_permission_bindings AS binding
  ON binding.role_id=role.id AND binding.state='active'
 LEFT JOIN iam_permissions AS permission ON permission.id=binding.permission_id
 LEFT JOIN iam_resources AS resource ON resource.id=permission.resource_id
 WHERE role.name IN (SELECT role_name FROM expected)
 GROUP BY role.name
)
SELECT
 (SELECT pg_catalog.sum(changelog_count) FROM actual),
 (SELECT pg_catalog.sum(usage_count) FROM actual),
 (SELECT pg_catalog.count(*) FROM expected
  JOIN actual ON actual.name=expected.role_name
  WHERE actual.changelog_count<>expected.changelog_count
     OR actual.usage_count<>expected.usage_count),
 (SELECT changelog_count+usage_count FROM actual WHERE name='BT Uzmanı'),
 (SELECT pg_catalog.count(*) FROM iam_role_permission_bindings AS binding
  JOIN iam_roles AS role ON role.id=binding.role_id
  JOIN iam_permissions AS permission ON permission.id=binding.permission_id
  JOIN iam_resources AS resource ON resource.id=permission.resource_id
  WHERE role.is_protected AND resource.resource_key IN
   ('changelog','analytics.usage'))`).Scan(
		&changelog, &usage, &mismatches, &unrelated, &protected,
	)
	if err != nil || changelog != 6 || usage != 2 || mismatches != 0 ||
		unrelated != 0 || protected != 0 {
		t.Fatalf("tenant observability grants=%d/%d mismatch=%d unrelated=%d protected=%d: %v",
			changelog, usage, mismatches, unrelated, protected, err)
	}
}

func assertIAMTenantObservabilityClock(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
) {
	t.Helper()
	var clock int64
	var changelogAllowed, usageAllowed bool
	err := transaction.QueryRowContext(ctx, `
SELECT (SELECT clock_value FROM iam_authorization_clock WHERE singleton),
 iam_user_has_permission_v1($1,'019fb202-0000-8000-8000-000000000201'),
 iam_user_has_permission_v1($1,'019fb202-0000-8000-8000-000000000202')`,
		iamUserOneID,
	).Scan(&clock, &changelogAllowed, &usageAllowed)
	if err != nil || clock != 1 || !changelogAllowed || !usageAllowed {
		t.Fatalf("tenant observability clock/tenant-all=%d/%t/%t: %v",
			clock, changelogAllowed, usageAllowed, err)
	}
}
