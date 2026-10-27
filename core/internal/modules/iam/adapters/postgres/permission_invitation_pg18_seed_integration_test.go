//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func seedIAMPermissionInvitation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
	defer connection.Close()
	for _, statement := range []string{
		seedIAMPermissionInvitationUsers,
		seedIAMPermissionInvitationCatalog,
	} {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("seed permission/invitation in %s: %v", tenant.schema, err)
		}
	}
	if _, err := transaction.ExecContext(
		ctx, seedIAMPermissionInvitationRole, tenant.id,
	); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("seed invitation role in %s: %v", tenant.schema, err)
	}
	if _, err := transaction.ExecContext(
		ctx, seedIAMPermissionInvitationBinding,
	); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("seed permission binding in %s: %v", tenant.schema, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit permission/invitation seed in %s: %v", tenant.schema, err)
	}
	seedIAMPermissionInvitationExpiryFixtures(t, ctx, database, tenant)
}

const seedIAMPermissionInvitationUsers = `
UPDATE iam_users
SET authorization_version=1, revision=revision+1,
    updated_at='2026-07-16T00:20:00Z';
INSERT INTO iam_user_identifiers (
    id, user_id, kind, display_value, normalized_value, normalized_sha256,
    state, revision, verified_at, activated_at, tombstoned_at,
    cooldown_until, retired_at, created_at, updated_at
) VALUES (
    'b0010000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000003', 'email_v1',
    'invitee@example.com', 'invitee@example.com',
    pg_catalog.decode(pg_catalog.repeat('81',32),'hex'), 'active', 1,
    '2026-07-16T00:05:00Z', '2026-07-16T00:06:00Z', NULL, NULL, NULL,
    '2026-07-16T00:00:00Z', '2026-07-16T00:06:00Z'
);`

const seedIAMPermissionInvitationCatalog = `
INSERT INTO iam_resources (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES
    ('a0010000-0000-4000-8000-000000000001','iam.permission_requests','iam',
     'request','Permission requests',NULL,true,'active',1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL),
    ('b0020000-0000-4000-8000-000000000001','iam.invitations','iam',
     'request','Invitations',NULL,true,'active',1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL);
INSERT INTO iam_actions (
    id, action_key, display_name, description, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES
    ('a0020000-0000-4000-8000-000000000001','manage','Manage',NULL,true,
     'active',1,'2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL),
    ('a0020000-0000-4000-8000-000000000002','review','Review',NULL,true,
     'active',1,'2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL);
INSERT INTO iam_permissions (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES
    ('a0030000-0000-4000-8000-000000000001','exact',
     'a0010000-0000-4000-8000-000000000001',
     'a0020000-0000-4000-8000-000000000001',true,'active',1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL),
    ('a0030000-0000-4000-8000-000000000002','exact',
     'a0010000-0000-4000-8000-000000000001',
     'a0020000-0000-4000-8000-000000000002',true,'active',1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL),
    ('b0030000-0000-4000-8000-000000000001','exact',
     'b0020000-0000-4000-8000-000000000001',
     'a0020000-0000-4000-8000-000000000001',true,'active',1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL);`

const seedIAMPermissionInvitationRole = `
INSERT INTO iam_roles (
    id, scope_kind, scope_id, organization_id, name, name_key,
    description, status, is_protected, is_default, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, retired_at
) VALUES (
    'b0040000-0000-4000-8000-000000000001','tenant',$1,NULL,
    'Invited operator',pg_catalog.convert_to('Invited operator','UTF8'),
    NULL,'active',false,false,'role',
    '41000000-0000-4000-8000-000000000003','{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('82',32),'hex'),1,
    '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL
);`

const seedIAMPermissionInvitationBinding = `
INSERT INTO iam_user_permission_bindings (
    id,user_id,permission_id,permission_match_kind,state,
    revision,created_at,updated_at,activated_at,revoked_at
) VALUES (
    'a0040000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000003',
    '93000000-0000-4000-8000-000000000001','exact','active',1,
    '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',
    '2026-07-16T00:20:00Z',NULL
);`
