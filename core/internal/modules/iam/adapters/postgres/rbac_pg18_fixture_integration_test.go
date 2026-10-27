//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	iamRBACResourceID        = "91000000-0000-4000-8000-000000000001"
	iamRBACReadActionID      = "92000000-0000-4000-8000-000000000001"
	iamRBACWriteActionID     = "92000000-0000-4000-8000-000000000002"
	iamRBACReadPermissionID  = "93000000-0000-4000-8000-000000000001"
	iamRBACWritePermissionID = "93000000-0000-4000-8000-000000000002"
	iamRBACTenantAllID       = "93000000-0000-4000-8000-000000000003"
)

const insertIAMRBACResource = `
INSERT INTO iam_resources (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    '91000000-0000-4000-8000-000000000001', 'iam.users', 'iam',
    'system_resource', 'Users', NULL, true, 'active', 1,
    '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL
)`

const insertIAMRBACActions = `
INSERT INTO iam_actions (
    id, action_key, display_name, description, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES
    ('92000000-0000-4000-8000-000000000001', 'read', 'Read', NULL, false,
     'active', 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL),
    ('92000000-0000-4000-8000-000000000002', 'write', 'Write', NULL, true,
     'active', 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL)`

const insertIAMRBACPermissions = `
INSERT INTO iam_permissions (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES
    ('93000000-0000-4000-8000-000000000001', 'exact',
     '91000000-0000-4000-8000-000000000001',
     '92000000-0000-4000-8000-000000000001', false,
     'active', 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL),
    ('93000000-0000-4000-8000-000000000002', 'exact',
     '91000000-0000-4000-8000-000000000001',
     '92000000-0000-4000-8000-000000000002', true,
     'active', 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL),
    ('93000000-0000-4000-8000-000000000003', 'tenant_all', NULL, NULL, true,
     'active', 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z', NULL)`

const insertIAMAuthorizationClock = `
INSERT INTO iam_authorization_clock (
    singleton, clock_value, revision, created_at, updated_at
) VALUES (true, 0, 1, '2026-07-16T00:10:00Z', '2026-07-16T00:10:00Z')`

func seedIAMRBACCatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, insertIAMRBACResource)
	mustExecIAMTenant(t, ctx, database, schema, insertIAMRBACActions)
	mustExecIAMTenant(t, ctx, database, schema, insertIAMRBACPermissions)
	mustExecIAMTenant(t, ctx, database, schema, insertIAMAuthorizationClock)
}
