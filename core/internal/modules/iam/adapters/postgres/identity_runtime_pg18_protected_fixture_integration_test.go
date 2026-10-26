//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMIdentitySecondProtectedAssignment = `
INSERT INTO iam_user_role_assignments (
    id, user_id, role_id, role_scope_kind, role_scope_id,
    organization_id, membership_id, state, revision,
    created_at, updated_at, revoked_at
) VALUES (
    '95000000-0000-4000-8000-000000000002',
    '51000000-0000-4000-8000-000000000002',
    '94000000-0000-4000-8000-000000000001', 'tenant', $1,
    NULL, NULL, 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL
)`

const insertIAMIdentityProtectedPolicy = `
INSERT INTO iam_protected_admin_policies (
    singleton, role_id, role_is_protected, role_status, role_scope_kind,
    tenant_all_permission_id, permission_match_kind,
    permission_is_protected, permission_state,
    role_permission_binding_id, role_binding_state,
    minimum_active_assignments, policy_state, authorization_epoch,
    revision, created_at, updated_at
) VALUES (
    true, '94000000-0000-4000-8000-000000000001', true, 'active', 'tenant',
    '93000000-0000-4000-8000-000000000003', 'tenant_all', true, 'active',
    '96000000-0000-4000-8000-000000000001', 'active',
    2, 'active', 1, 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z'
)`

func seedIAMIdentityProtectedPolicy(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
	defer connection.Close()
	for _, statement := range []string{
		insertIAMRBACProtectedRole,
		insertIAMRBACProtectedAssignment,
		insertIAMIdentitySecondProtectedAssignment,
	} {
		if _, err := transaction.ExecContext(ctx, statement, tenant.id); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("seed identity protected assignment in %s: %v", tenant.schema, err)
		}
	}
	for _, statement := range []string{
		insertIAMRBACProtectedBinding,
		insertIAMIdentityProtectedPolicy,
	} {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("seed identity protected policy in %s: %v", tenant.schema, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit identity protected policy in %s: %v", tenant.schema, err)
	}
}
