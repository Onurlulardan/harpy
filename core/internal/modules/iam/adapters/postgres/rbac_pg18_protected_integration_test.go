//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const revokeIAMProtectedAssignment = `
UPDATE iam_user_role_assignments
SET state = 'revoked', revision = 2,
    updated_at = '2026-07-16T00:12:00Z', revoked_at = '2026-07-16T00:12:00Z'
WHERE id = $1`

func assertIAMRBACProtectedAdminGuards(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	assignmentErr := rejectIAMTenant(t, ctx, database, tenant.schema,
		revokeIAMProtectedAssignment, iamRBACProtectedAssignmentID)
	assertIAMConstraint(t, assignmentErr, "23514", "iam_protected_admin_minimum_guard")
	userErr := rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_users SET status = 'suspended', revision = 2,
    updated_at = '2026-07-16T00:12:00Z'
WHERE id = $1`, iamUserOneID)
	assertIAMConstraint(t, userErr, "23514", "iam_protected_admin_minimum_guard")
	policyErr := rejectIAMTenant(t, ctx, database, tenant.schema,
		"DELETE FROM iam_protected_admin_policies WHERE singleton")
	assertIAMConstraint(t, policyErr, "23514",
		"iam_protected_admin_policies_forbidden_mutation_guard")
	bindingErr := rejectIAMTenant(t, ctx, database, tenant.schema, `
UPDATE iam_role_permission_bindings
SET state = 'revoked', revision = 2,
    updated_at = '2026-07-16T00:12:00Z', revoked_at = '2026-07-16T00:12:00Z'
WHERE id = $1`, iamRBACProtectedBindingID)
	assertIAMConstraint(t, bindingErr, "23503", "iam_protected_admin_policies_binding_fk")
	assertIAMRBACOrphanWildcardRejected(t, ctx, database, tenant)
	assertIAMRBACProtectedAdminRace(t, ctx, database, tenant)
}

func assertIAMRBACOrphanWildcardRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMRBACProtectedRoleProbe,
		"98000000-0000-4000-8000-000000000001", tenant.id, "Orphan Protected")
	err := rejectIAMTenantCommit(t, ctx, database, tenant.schema,
		insertIAMRBACRoleBinding,
		"98000000-0000-4000-8000-000000000002",
		"98000000-0000-4000-8000-000000000001", true,
		iamRBACTenantAllID, "tenant_all")
	assertIAMConstraint(t, err, "23514",
		"iam_role_permission_bindings_wildcard_policy_guard")
}

const insertIAMRBACProtectedRoleProbe = `
INSERT INTO iam_roles (
    id, scope_kind, scope_id, organization_id, name, name_key,
    description, status, is_protected, is_default, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, retired_at
) VALUES ($1, 'tenant', $2, NULL, $3, pg_catalog.convert_to($3, 'UTF8'),
    NULL, 'active', true, false, 'role',
    '41000000-0000-4000-8000-000000000003', '{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('98', 32), 'hex'), 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL)`

func rejectIAMTenantCommit(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) error {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	if _, err := transaction.ExecContext(ctx, statement, arguments...); err != nil {
		_ = transaction.Rollback()
		return err
	}
	err := transaction.Commit()
	if err == nil {
		t.Fatalf("IAM transaction unexpectedly committed in %s", schema)
	}
	return err
}

func assertIAMRBACProtectedAdminRace(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_user_role_assignments (
    id, user_id, role_id, role_scope_kind, role_scope_id,
    organization_id, membership_id, state, revision,
    created_at, updated_at, revoked_at
) VALUES ('95000000-0000-4000-8000-000000000002', $1, $2, 'tenant', $3,
    NULL, NULL, 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL)`,
		iamUserTwoID, iamRBACProtectedRoleID, tenant.id)
	firstConnection, first := beginIAMTenantTx(t, ctx, database, tenant.schema)
	defer firstConnection.Close()
	secondConnection, second := beginIAMTenantTx(t, ctx, database, tenant.schema)
	defer secondConnection.Close()
	if _, err := first.ExecContext(ctx, revokeIAMProtectedAssignment,
		iamRBACProtectedAssignmentID); err != nil {
		t.Fatalf("stage first protected assignment revoke: %v", err)
	}
	result := make(chan error, 1)
	go func() {
		_, err := second.ExecContext(ctx, revokeIAMProtectedAssignment,
			"95000000-0000-4000-8000-000000000002")
		result <- err
	}()
	select {
	case err := <-result:
		_ = first.Rollback()
		_ = second.Rollback()
		t.Fatalf("concurrent protected revoke did not wait on policy lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	if err := first.Commit(); err != nil {
		_ = second.Rollback()
		t.Fatalf("commit first protected assignment revoke: %v", err)
	}
	err := <-result
	_ = second.Rollback()
	assertIAMConstraint(t, err, "23514", "iam_protected_admin_minimum_guard")
}
