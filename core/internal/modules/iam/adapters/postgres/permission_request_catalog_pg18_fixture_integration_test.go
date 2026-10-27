//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const seedIAMPermissionRequestCatalogRows = `
INSERT INTO iam_resources (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES
    ('d1000000-0000-4000-8000-000000000001','form.orders','forms','form',
     'Orders','Requestable form',false,'active',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:00:00Z',NULL),
    ('d1000000-0000-4000-8000-000000000002','crm.contacts','crm','record',
     'Contacts','Requestable system resource',false,'active',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:00:00Z',NULL),
    ('d1000000-0000-4000-8000-000000000003','form.secret','forms','form',
     'Secret form','Protected resource',true,'active',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:00:00Z',NULL),
    ('d1000000-0000-4000-8000-000000000004','form.retired','forms','form',
     'Retired form','Inactive resource',false,'retired',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:01:00Z','2026-08-01T00:01:00Z');
INSERT INTO iam_permissions (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES
    ('d3000000-0000-4000-8000-000000000001','exact',
     'd1000000-0000-4000-8000-000000000001',
     '92000000-0000-4000-8000-000000000001',false,'active',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:00:00Z',NULL),
    ('d3000000-0000-4000-8000-000000000002','exact',
     'd1000000-0000-4000-8000-000000000002',
     '92000000-0000-4000-8000-000000000001',false,'active',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:00:00Z',NULL),
    ('d3000000-0000-4000-8000-000000000003','exact',
     'd1000000-0000-4000-8000-000000000003',
     '92000000-0000-4000-8000-000000000001',false,'active',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:00:00Z',NULL),
    ('d3000000-0000-4000-8000-000000000004','exact',
     'd1000000-0000-4000-8000-000000000001',
     '92000000-0000-4000-8000-000000000002',false,'active',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:00:00Z',NULL),
    ('d3000000-0000-4000-8000-000000000005','exact',
     'd1000000-0000-4000-8000-000000000004',
     '92000000-0000-4000-8000-000000000001',false,'retired',1,
     '2026-08-01T00:00:00Z','2026-08-01T00:01:00Z','2026-08-01T00:01:00Z');
INSERT INTO iam_user_permission_bindings (
    id,user_id,permission_id,permission_match_kind,state,
    revision,created_at,updated_at,activated_at,revoked_at
) VALUES (
    'd4000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000002',
    'd3000000-0000-4000-8000-000000000001','exact','active',1,
    '2026-08-01T00:00:00Z','2026-08-01T00:00:00Z',
    '2026-08-01T00:00:00Z',NULL
);`

func seedIAMPermissionRequestCatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, seedIAMPermissionRequestCatalogRows)
}
