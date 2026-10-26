//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMRBACHTTPTypedEffective(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	role := mustCreateIAMRBACHTTPAssignment(
		t, ctx, database, "t_iam_b",
		"b1000000-0000-4000-8000-000000000004",
		"b2000000-0000-4000-8000-000000000002",
		"ROLE", iamRBACProtectedRoleID, iamRBACReadActionID,
		"b3000000-0000-4000-8000-000000000004",
		"b4000000-0000-4000-8000-000000000004", 0xd4,
	)
	if role != "b2000000-0000-4000-8000-000000000002|1|2|false|true" {
		t.Fatalf("role assignment = %s", role)
	}
	organization := mustCreateIAMRBACHTTPAssignment(
		t, ctx, database, "t_iam_b",
		"b1000000-0000-4000-8000-000000000005",
		"b2000000-0000-4000-8000-000000000003",
		"ORGANIZATION", iamRBACHTTPOrganizationID, iamRBACWriteActionID,
		"b3000000-0000-4000-8000-000000000005",
		"b4000000-0000-4000-8000-000000000005", 0xe5,
	)
	if organization != "b2000000-0000-4000-8000-000000000003|1|3|false|true" {
		t.Fatalf("organization assignment = %s", organization)
	}
	assertIAMRBACHTTPTypedBindings(t, ctx, database)
	assertIAMRBACHTTPEffectiveViews(t, ctx, database)
}

func assertIAMRBACHTTPTypedBindings(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var roleExact, organizationExact, forbiddenWildcards int
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_b", `
SELECT
 (SELECT pg_catalog.count(*) FROM iam_role_permission_bindings
  WHERE role_id='94000000-0000-4000-8000-000000000001'
    AND permission_id='93000000-0000-4000-8000-000000000001'
    AND permission_match_kind='exact' AND state='active'),
 (SELECT pg_catalog.count(*) FROM iam_organization_permission_bindings
  WHERE organization_id='a1000000-0000-4000-8000-000000000001'
    AND permission_id='93000000-0000-4000-8000-000000000002'
    AND permission_match_kind='exact' AND state='active'),
 (SELECT pg_catalog.count(*) FROM iam_user_permission_bindings
  WHERE permission_match_kind<>'exact') +
 (SELECT pg_catalog.count(*) FROM iam_organization_permission_bindings
  WHERE permission_match_kind<>'exact')`,
		&roleExact, &organizationExact, &forbiddenWildcards)
	if roleExact != 1 || organizationExact != 1 || forbiddenWildcards != 0 {
		t.Fatalf("typed bindings role/org/wildcards = %d/%d/%d",
			roleExact, organizationExact, forbiddenWildcards)
	}
}

func assertIAMRBACHTTPEffectiveViews(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var roleCount, organizationCount, userCount int
	query := `SELECT pg_catalog.count(*) FROM iam_rbac_effective_permissions_v1(
  pg_catalog.current_schema()::name, $1, $2, 100,
  '51000000-0000-4000-8000-000000000001',
 	  '93000000-0000-4000-8000-000000000001')`
	roleCount = mustIAMRBACHTTPEffectiveCount(
		t, ctx, database, "t_iam_b", query, "ROLE", iamRBACProtectedRoleID,
	)
	organizationCount = mustIAMRBACHTTPEffectiveCount(
		t, ctx, database, "t_iam_b", query, "ORGANIZATION", iamRBACHTTPOrganizationID,
	)
	userCount = mustIAMRBACHTTPEffectiveCount(
		t, ctx, database, "t_iam_b", query, "USER", iamUserOneID,
	)
	if roleCount < 2 || organizationCount != 1 || userCount < 3 {
		t.Fatalf("effective role/org/user = %d/%d/%d",
			roleCount, organizationCount, userCount)
	}
}

func mustIAMRBACHTTPEffectiveCount(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	query string,
	target string,
	targetID string,
) int {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatalf("begin effective %s permissions: %v", target, err)
	}
	defer connection.Close()
	var count int
	if err = transaction.QueryRowContext(ctx, query, target, targetID).Scan(&count); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("read effective %s permissions: %v", target, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return count
}
