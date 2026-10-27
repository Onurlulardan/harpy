//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const insertIAMRBACUserBinding = `
INSERT INTO iam_user_permission_bindings (
    id, user_id, permission_id, permission_match_kind, state, revision,
    created_at, updated_at, activated_at, revoked_at
) VALUES ($1, $2, $3, $4, 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z',
    '2026-07-16T00:10:00Z', NULL)`

const insertIAMRBACRoleBinding = `
INSERT INTO iam_role_permission_bindings (
    id, role_id, role_is_protected, permission_id, permission_match_kind,
    state, revision, created_at, updated_at, activated_at, revoked_at
) VALUES ($1, $2, $3, $4, $5, 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z',
    '2026-07-16T00:10:00Z', NULL)`

const insertIAMRBACOrganizationBinding = `
INSERT INTO iam_organization_permission_bindings (
    id, organization_id, permission_id, permission_match_kind,
    state, revision, created_at, updated_at, activated_at, revoked_at
) VALUES ($1, $2, $3, $4, 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z',
    '2026-07-16T00:10:00Z', NULL)`

func assertIAMRBACRelationalConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	assertIAMRBACTypedBindings(t, ctx, database, schema)
	wildcardUser := rejectIAMTenant(t, ctx, database, schema,
		insertIAMRBACUserBinding,
		"97000000-0000-4000-8000-000000000011", iamUserTwoID,
		iamRBACTenantAllID, "tenant_all")
	assertIAMConstraint(t, wildcardUser, "23514", "iam_user_permission_bindings_exact_check")
	wildcardOrganization := rejectIAMTenant(t, ctx, database, schema,
		insertIAMRBACOrganizationBinding,
		"97000000-0000-4000-8000-000000000012",
		"97000000-0000-4000-8000-000000000002", iamRBACTenantAllID, "tenant_all")
	assertIAMConstraint(t, wildcardOrganization, "23514",
		"iam_organization_permission_bindings_exact_check")
	assertIAMRBACRetiredCatalogRejectsGrant(t, ctx, database, schema)
	removeErr := rejectIAMTenant(t, ctx, database, schema,
		"DELETE FROM iam_actions WHERE id = $1", iamRBACReadActionID)
	assertIAMConstraint(t, removeErr, "23514", "iam_actions_forbidden_mutation_guard")
	clockErr := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_authorization_clock
SET clock_value = 2, revision = 2, updated_at = '2026-07-16T00:11:00Z'
WHERE singleton`)
	assertIAMConstraint(t, clockErr, "23514", "iam_authorization_clock_step_guard")
	assertIAMRBACReceiptIsImmutable(t, ctx, database, schema)
}

func assertIAMRBACTypedBindings(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	organizationID := "97000000-0000-4000-8000-000000000002"
	mustExecIAMTenant(t, ctx, database, schema, insertIAMOrganization,
		organizationID, "RBAC Organization", "rbac-organization", nil)
	mustExecIAMTenant(t, ctx, database, schema, insertIAMRBACUserBinding,
		"97000000-0000-4000-8000-000000000003", iamUserOneID,
		iamRBACReadPermissionID, "exact")
	mustExecIAMTenant(t, ctx, database, schema, insertIAMRBACRoleBinding,
		"97000000-0000-4000-8000-000000000004", iamRBACProtectedRoleID, true,
		iamRBACReadPermissionID, "exact")
	mustExecIAMTenant(t, ctx, database, schema, insertIAMRBACOrganizationBinding,
		"97000000-0000-4000-8000-000000000005", organizationID,
		iamRBACReadPermissionID, "exact")
}

func assertIAMRBACRetiredCatalogRejectsGrant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_resources (
    id, resource_key, owner_domain, resource_kind, display_name, description,
    is_protected, state, revision, created_at, updated_at, retired_at
) VALUES ('97000000-0000-4000-8000-000000000006', 'iam.retired-probe',
    'iam', 'system_resource', 'Retired Probe', NULL, false, 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL)`)
	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_resources SET state = 'retired', revision = 2,
    updated_at = '2026-07-16T00:11:00Z', retired_at = '2026-07-16T00:11:00Z'
WHERE id = '97000000-0000-4000-8000-000000000006'`)
	err := rejectIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_permissions (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES ('97000000-0000-4000-8000-000000000007', 'exact',
    '97000000-0000-4000-8000-000000000006', $1, false, 'active', 1,
    '2026-07-16T00:11:00Z', '2026-07-16T00:11:00Z', NULL)`, iamRBACReadActionID)
	assertIAMConstraint(t, err, "23514", "iam_permissions_active_catalog_guard")
}

func assertIAMRBACReceiptIsImmutable(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_authorization_operation_receipts (
    operation_id, command_key, request_sha256, clock_value, result_code,
    result_id, result_revision, result_sha256, created_at
) VALUES ('97000000-0000-4000-8000-000000000008', 'iam.permission.grant',
    pg_catalog.decode(pg_catalog.repeat('a1', 32), 'hex'), 1, 'grant.applied',
    $1, 1, pg_catalog.decode(pg_catalog.repeat('a2', 32), 'hex'),
    '2026-07-16T00:11:00Z')`, iamRBACReadPermissionID)
	err := rejectIAMTenant(t, ctx, database, schema, `
UPDATE iam_authorization_operation_receipts SET result_revision = 2
WHERE operation_id = '97000000-0000-4000-8000-000000000008'`)
	assertIAMConstraint(t, err, "23514",
		"iam_authorization_operation_receipts_forbidden_mutation_guard")
}
