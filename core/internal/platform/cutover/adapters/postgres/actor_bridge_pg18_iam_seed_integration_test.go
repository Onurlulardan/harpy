//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	actorBridgeUserID       = "51000000-0000-4000-8000-000000000001"
	actorBridgePermissionID = "93000000-0000-4000-8000-000000000001"
)

func seedActorBridgeIAMTenant(
	t *testing.T, ctx context.Context, migrator *pgx.Conn, tenant actorBridgeTenantFixture,
) {
	t.Helper()
	if _, err := migrator.Exec(ctx, actorBridgeIAMSeedSQL,
		pgx.QueryExecModeSimpleProtocol, tenant.id); err != nil {
		t.Fatalf("seed actor bridge IAM tenant %s: %v", tenant.schema, err)
	}
	if _, err := migrator.Exec(ctx, `
CREATE TABLE actor_bridge_business_rows (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL
);
REVOKE ALL ON TABLE actor_bridge_business_rows FROM PUBLIC;
GRANT SELECT, INSERT ON TABLE actor_bridge_business_rows TO harpy_runtime;`); err != nil {
		t.Fatalf("create actor bridge business table in %s: %v", tenant.schema, err)
	}
	if _, err := migrator.Exec(ctx, "GRANT USAGE ON SCHEMA "+
		pgx.Identifier{tenant.schema}.Sanitize()+" TO harpy_runtime"); err != nil {
		t.Fatalf("grant actor bridge schema usage in %s: %v", tenant.schema, err)
	}
}

const actorBridgeIAMSeedSQL = `
INSERT INTO iam_profile_schema_versions (
    id, resource_kind, version, state, definition_sha256,
    activated_at, retired_at, created_at, updated_at, revision
) VALUES
    ('41000000-0000-4000-8000-000000000001','user',1,'active',
     pg_catalog.decode(pg_catalog.repeat('11',32),'hex'),
     '2026-07-16T00:00:00Z',NULL,'2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',1),
    ('41000000-0000-4000-8000-000000000003','role',1,'active',
     pg_catalog.decode(pg_catalog.repeat('13',32),'hex'),
     '2026-07-16T00:00:00Z',NULL,'2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',1);
INSERT INTO iam_users (
    id, identity_kind, status, profile_resource_kind, profile_schema_version_id,
    profile_values, profile_sha256, token_version, authorization_version,
    revision, created_at, updated_at, tombstoned_at
) VALUES
    (
    '51000000-0000-4000-8000-000000000001','human','active','user',
    '41000000-0000-4000-8000-000000000001','{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('21',32),'hex'),1,1,1,
    '2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL
    ),
    (
    '51000000-0000-4000-8000-000000000002','human','active','user',
    '41000000-0000-4000-8000-000000000001','{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('22',32),'hex'),1,1,1,
    '2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL
    );
INSERT INTO iam_resources (
    id, resource_key, owner_domain, resource_kind, display_name, description,
    is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    '91000000-0000-4000-8000-000000000001','forms.records','forms',
    'system_resource','Form records',NULL,true,'active',1,
    '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL
);
INSERT INTO iam_actions (
    id, action_key, display_name, description, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES (
    '92000000-0000-4000-8000-000000000001','read','Read',NULL,false,
    'active',1,'2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL
);
INSERT INTO iam_permissions (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES
    ('93000000-0000-4000-8000-000000000001','exact',
     '91000000-0000-4000-8000-000000000001',
     '92000000-0000-4000-8000-000000000001',false,'active',1,
     '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL),
    ('93000000-0000-4000-8000-000000000003','tenant_all',
     NULL,NULL,true,'active',1,
     '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL);
INSERT INTO iam_authorization_clock (
    singleton, clock_value, revision, created_at, updated_at
) VALUES (true,0,1,'2026-07-16T00:10:00Z','2026-07-16T00:10:00Z');
INSERT INTO iam_roles (
    id, scope_kind, scope_id, organization_id, name, name_key, description,
    status, is_protected, is_default, profile_resource_kind,
    profile_schema_version_id, profile_values, profile_sha256,
    revision, created_at, updated_at, retired_at
) VALUES (
    '94000000-0000-4000-8000-000000000001','tenant',$1,NULL,
    'Protected Tenant Admin',pg_catalog.convert_to('Protected Tenant Admin','UTF8'),
    NULL,'active',true,false,'role','41000000-0000-4000-8000-000000000003',
    '{}'::jsonb,pg_catalog.decode(pg_catalog.repeat('91',32),'hex'),1,
    '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL
);
INSERT INTO iam_user_role_assignments (
    id, user_id, role_id, role_scope_kind, role_scope_id, organization_id,
    membership_id, state, revision, created_at, updated_at, revoked_at
) VALUES (
    '95000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000002',
    '94000000-0000-4000-8000-000000000001','tenant',$1,
    NULL,NULL,'active',1,'2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL
);
INSERT INTO iam_role_permission_bindings (
    id, role_id, role_is_protected, permission_id, permission_match_kind,
    state, revision, created_at, updated_at, activated_at, revoked_at
) VALUES (
    '96000000-0000-4000-8000-000000000001',
    '94000000-0000-4000-8000-000000000001',true,
    '93000000-0000-4000-8000-000000000003','tenant_all','active',1,
    '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL
);
INSERT INTO iam_protected_admin_policies (
    singleton, role_id, role_is_protected, role_status, role_scope_kind,
    tenant_all_permission_id, permission_match_kind, permission_is_protected,
    permission_state, role_permission_binding_id, role_binding_state,
    minimum_active_assignments, policy_state, authorization_epoch,
    revision, created_at, updated_at
) VALUES (
    true,'94000000-0000-4000-8000-000000000001',true,'active','tenant',
    '93000000-0000-4000-8000-000000000003','tenant_all',true,'active',
    '96000000-0000-4000-8000-000000000001','active',1,'active',1,1,
    '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z'
);
INSERT INTO iam_user_permission_bindings (
    id, user_id, permission_id, permission_match_kind, state,
    revision, created_at, updated_at, activated_at, revoked_at
) VALUES (
    '97000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001','exact','active',1,
    '2026-07-16T00:10:00Z','2026-07-16T00:10:00Z','2026-07-16T00:10:00Z',NULL
);`
