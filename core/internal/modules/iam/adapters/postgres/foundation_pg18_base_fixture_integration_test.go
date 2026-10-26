//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	iamUserProfileSchemaID         = "41000000-0000-4000-8000-000000000001"
	iamOrganizationProfileSchemaID = "41000000-0000-4000-8000-000000000002"
	iamRoleProfileSchemaID         = "41000000-0000-4000-8000-000000000003"
	iamUserOneID                   = "51000000-0000-4000-8000-000000000001"
	iamUserTwoID                   = "51000000-0000-4000-8000-000000000002"
	iamUserThreeID                 = "51000000-0000-4000-8000-000000000003"
	iamUserOneConsoleModeID        = "52000000-0000-4000-8000-000000000001"
	iamUserTwoPortalModeID         = "52000000-0000-4000-8000-000000000002"
	iamUserThreeConsoleModeID      = "52000000-0000-4000-8000-000000000003"
	iamUserThreePortalModeID       = "52000000-0000-4000-8000-000000000004"
	iamUserOnePortalModeID         = "52000000-0000-4000-8000-000000000005"
)

const insertIAMProfileRoots = `
INSERT INTO iam_profile_schema_versions (
    id, resource_kind, version, state, definition_sha256,
    activated_at, retired_at, created_at, updated_at, revision
) VALUES
    ('41000000-0000-4000-8000-000000000001', 'user', 1, 'active',
     pg_catalog.decode(pg_catalog.repeat('11', 32), 'hex'),
     '2026-07-16T00:00:00Z', NULL,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', 1),
    ('41000000-0000-4000-8000-000000000002', 'organization', 1, 'active',
     pg_catalog.decode(pg_catalog.repeat('12', 32), 'hex'),
     '2026-07-16T00:00:00Z', NULL,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', 1),
    ('41000000-0000-4000-8000-000000000003', 'role', 1, 'active',
     pg_catalog.decode(pg_catalog.repeat('13', 32), 'hex'),
     '2026-07-16T00:00:00Z', NULL,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', 1);`

const insertIAMBaseUsers = `
INSERT INTO iam_users (
    id, identity_kind, status, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    token_version, authorization_version, revision, created_at,
    updated_at, tombstoned_at
) VALUES
    ('51000000-0000-4000-8000-000000000001', 'human', 'active', 'user',
     '41000000-0000-4000-8000-000000000001', '{}'::jsonb,
     pg_catalog.decode(pg_catalog.repeat('21', 32), 'hex'), 1, 0, 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL),
    ('51000000-0000-4000-8000-000000000002', 'human', 'active', 'user',
     '41000000-0000-4000-8000-000000000001', '{}'::jsonb,
     pg_catalog.decode(pg_catalog.repeat('22', 32), 'hex'), 1, 0, 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL),
    ('51000000-0000-4000-8000-000000000003', 'human', 'active', 'user',
     '41000000-0000-4000-8000-000000000001', '{}'::jsonb,
     pg_catalog.decode(pg_catalog.repeat('23', 32), 'hex'), 1, 0, 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL);`

const insertIAMBaseAccessModes = `
INSERT INTO iam_user_access_modes (
    id, user_id, mode, state, revision, created_at, updated_at,
    activated_at, suspended_at, retired_at
) VALUES
    ('52000000-0000-4000-8000-000000000001',
     '51000000-0000-4000-8000-000000000001',
     'operator_console', 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
     '2026-07-16T00:00:00Z', NULL, NULL),
    ('52000000-0000-4000-8000-000000000002',
     '51000000-0000-4000-8000-000000000002',
     'portal', 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
     '2026-07-16T00:00:00Z', NULL, NULL),
    ('52000000-0000-4000-8000-000000000003',
     '51000000-0000-4000-8000-000000000003',
     'operator_console', 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
     '2026-07-16T00:00:00Z', NULL, NULL),
    ('52000000-0000-4000-8000-000000000004',
     '51000000-0000-4000-8000-000000000003',
     'portal', 'active', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z',
     '2026-07-16T00:00:00Z', NULL, NULL),
    ('52000000-0000-4000-8000-000000000005',
     '51000000-0000-4000-8000-000000000001',
     'portal', 'suspended', 1,
     '2026-07-16T00:00:00Z', '2026-07-16T00:01:00Z',
     NULL, '2026-07-16T00:01:00Z', NULL);`

func seedIAMFoundationBase(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMProfileRoots)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMBaseUsers)
		mustExecIAMTenant(t, ctx, database, tenant.schema, insertIAMBaseAccessModes)
	}
}
