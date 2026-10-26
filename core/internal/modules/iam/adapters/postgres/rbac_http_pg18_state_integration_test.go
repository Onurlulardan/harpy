//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMRBACHTTPBindingReplacement(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var revokedRead, activeWrite, activeMembers, revokedMembers int
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a", `
SELECT
    pg_catalog.count(*) FILTER (
        WHERE permission_id='93000000-0000-4000-8000-000000000001'
          AND state='revoked'
    ),
    pg_catalog.count(*) FILTER (
        WHERE permission_id='93000000-0000-4000-8000-000000000002'
          AND state='active'
    )
FROM iam_user_permission_bindings
WHERE user_id='51000000-0000-4000-8000-000000000002'`,
		&revokedRead, &activeWrite)
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a", `
SELECT pg_catalog.count(*) FILTER (WHERE state='active'),
       pg_catalog.count(*) FILTER (WHERE state='revoked')
FROM iam_rbac_assignment_members
WHERE group_id='b2000000-0000-4000-8000-000000000001'`,
		&activeMembers, &revokedMembers)
	if revokedRead != 1 || activeWrite != 1 || activeMembers != 1 || revokedMembers != 1 {
		t.Fatalf("replacement read/write/members = %d/%d/%d/%d",
			revokedRead, activeWrite, activeMembers, revokedMembers)
	}
}

func assertIAMRBACHTTPRetiredState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var state string
	var revision int64
	var activeBindings, activeMembers, audits, receipts int
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a", `
SELECT state, revision FROM iam_rbac_assignment_groups
WHERE id='b2000000-0000-4000-8000-000000000001'`, &state, &revision)
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a", `
SELECT
 (SELECT pg_catalog.count(*) FROM iam_user_permission_bindings
  WHERE user_id='51000000-0000-4000-8000-000000000002' AND state='active'),
 (SELECT pg_catalog.count(*) FROM iam_rbac_assignment_members
  WHERE group_id='b2000000-0000-4000-8000-000000000001' AND state='active'),
 (SELECT pg_catalog.count(*) FROM iam_rbac_change_audit
  WHERE aggregate_id='b2000000-0000-4000-8000-000000000001'),
 (SELECT pg_catalog.count(*) FROM iam_authorization_operation_receipts
  WHERE result_id='b2000000-0000-4000-8000-000000000001')`,
		&activeBindings, &activeMembers, &audits, &receipts)
	effective := mustIAMRBACHTTPEffectiveCount(t, ctx, database, "t_iam_a", `
SELECT pg_catalog.count(*) FROM iam_rbac_effective_permissions_v1(
  pg_catalog.current_schema()::name, $1, $2, 100,
  '51000000-0000-4000-8000-000000000001',
  '93000000-0000-4000-8000-000000000001')`, "USER", iamUserTwoID)
	if state != "retired" || revision != 3 || activeBindings != 0 ||
		activeMembers != 0 || effective != 0 || audits != 3 || receipts != 3 {
		t.Fatalf("retired state=%s rev=%d bindings=%d members=%d effective=%d audit=%d receipt=%d",
			state, revision, activeBindings, activeMembers, effective, audits, receipts)
	}
}
