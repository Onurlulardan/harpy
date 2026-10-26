//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	iamRBACProtectedRoleID       = "94000000-0000-4000-8000-000000000001"
	iamRBACProtectedAssignmentID = "95000000-0000-4000-8000-000000000001"
	iamRBACProtectedBindingID    = "96000000-0000-4000-8000-000000000001"
)

const insertIAMRBACProtectedRole = `
INSERT INTO iam_roles (
    id, scope_kind, scope_id, organization_id, name, name_key,
    description, status, is_protected, is_default, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, retired_at
) VALUES (
    '94000000-0000-4000-8000-000000000001', 'tenant', $1, NULL,
    'Protected Tenant Admin', pg_catalog.convert_to('Protected Tenant Admin', 'UTF8'),
    NULL, 'active', true, false, 'role',
    '41000000-0000-4000-8000-000000000003', '{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('91', 32), 'hex'), 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL
)`

const insertIAMRBACProtectedAssignment = `
INSERT INTO iam_user_role_assignments (
    id, user_id, role_id, role_scope_kind, role_scope_id,
    organization_id, membership_id, state, revision,
    created_at, updated_at, revoked_at
) VALUES (
    '95000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000001',
    '94000000-0000-4000-8000-000000000001', 'tenant', $1,
    NULL, NULL, 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL
)`

const insertIAMRBACProtectedBinding = `
INSERT INTO iam_role_permission_bindings (
    id, role_id, role_is_protected, permission_id, permission_match_kind,
    state, revision, created_at, updated_at, activated_at, revoked_at
) VALUES (
    '96000000-0000-4000-8000-000000000001',
    '94000000-0000-4000-8000-000000000001', true,
    '93000000-0000-4000-8000-000000000003', 'tenant_all',
    'active', 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z',
    '2026-07-16T00:10:00Z', NULL
)`

const insertIAMRBACProtectedPolicy = `
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
    1, 'active', 1, 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z'
)`

func seedIAMRBACProtectedPolicy(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
	defer connection.Close()
	for _, statement := range []string{
		insertIAMRBACProtectedRole, insertIAMRBACProtectedAssignment,
	} {
		if _, err := transaction.ExecContext(ctx, statement, tenant.id); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("seed protected policy in %s: %v", tenant.schema, err)
		}
	}
	for _, statement := range []string{
		insertIAMRBACProtectedBinding, insertIAMRBACProtectedPolicy,
	} {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("seed protected policy in %s: %v", tenant.schema, err)
		}
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit protected policy in %s: %v", tenant.schema, err)
	}
}
