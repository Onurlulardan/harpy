//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const ownerTransferFoundation = `
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA platform TO harpy_runtime;
CREATE SCHEMA t_owner AUTHORIZATION harpy_owner;
CREATE TABLE platform.tenants (id uuid PRIMARY KEY, schema_name name NOT NULL UNIQUE);
CREATE TABLE platform.environment_identity (
    singleton boolean PRIMARY KEY CHECK (singleton), environment_id uuid NOT NULL UNIQUE
);
CREATE TABLE platform.tenancy_deployment_policy (
    singleton boolean PRIMARY KEY CHECK (singleton), environment_id uuid NOT NULL UNIQUE,
    deployment_mode text NOT NULL, maximum_customer_tenants integer NOT NULL, created_at timestamptz NOT NULL
);
CREATE TABLE platform.tenancy_tenant_bindings (
    environment_id uuid NOT NULL, tenant_id uuid NOT NULL, purpose text NOT NULL,
    created_at timestamptz NOT NULL, PRIMARY KEY (environment_id,tenant_id)
);
INSERT INTO platform.tenants VALUES ('10000000-0000-4000-8000-000000000001','t_owner');
INSERT INTO platform.environment_identity VALUES (true,'30000000-0000-4000-8000-000000000001');
INSERT INTO platform.tenancy_deployment_policy VALUES
    (true,'30000000-0000-4000-8000-000000000001','onpremise',1,clock_timestamp());
INSERT INTO platform.tenancy_tenant_bindings VALUES
    ('30000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001',
     'onpremise_primary',clock_timestamp());
SET search_path TO t_owner,pg_catalog;
CREATE TABLE tenancy_schema_identity (
    singleton boolean PRIMARY KEY CHECK (singleton), tenant_id uuid NOT NULL UNIQUE
);
INSERT INTO tenancy_schema_identity VALUES (true,'10000000-0000-4000-8000-000000000001');
CREATE TABLE iam_users (
    id uuid PRIMARY KEY, identity_kind text NOT NULL, status text NOT NULL,
    token_version bigint NOT NULL, authorization_version bigint NOT NULL,
    revision bigint NOT NULL, created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL
);
CREATE TABLE iam_sessions (
    id uuid PRIMARY KEY, user_id uuid NOT NULL REFERENCES iam_users(id), state text NOT NULL,
    revision bigint NOT NULL, created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL,
    revoked_at timestamptz
);
INSERT INTO iam_users VALUES
    ('41000000-0000-4000-8000-000000000001','human','active',1,1,1,clock_timestamp(),clock_timestamp()),
    ('42000000-0000-4000-8000-000000000002','human','active',1,1,1,clock_timestamp(),clock_timestamp());
INSERT INTO iam_sessions VALUES
    ('43000000-0000-4000-8000-000000000001','41000000-0000-4000-8000-000000000001',
     'active',1,clock_timestamp(),clock_timestamp(),NULL),
    ('43000000-0000-4000-8000-000000000002','42000000-0000-4000-8000-000000000002',
     'active',1,clock_timestamp(),clock_timestamp(),NULL);
GRANT USAGE ON SCHEMA t_owner TO harpy_runtime;`

var ownerTransferPlatformMigrations = []string{
	"20260722020000_create_iam_platform_catalog",
	"20260722020001_create_iam_platform_authority",
	"20260722020002_create_iam_platform_policy",
	"20260722020003_guard_iam_platform_authority",
	"20260722020004_guard_iam_platform_policy",
	"20260722020005_guard_iam_platform_catalog",
	"20260722020006_seed_iam_platform_catalog",
	"20260722020007_seed_iam_platform_permissions",
	"20260722020008_create_iam_platform_context_v1",
	"20260722020009_create_iam_platform_operator_grant_apply_v1",
	"20260722020010_create_iam_platform_operator_revoke_apply_v1",
	"20260722020011_create_iam_platform_authorize_decide_v1",
	"20260722020012_create_iam_platform_owner_transfer_storage",
	"20260722020013_create_iam_platform_owner_transfer_compensations",
	"20260722020014_guard_iam_platform_owner_transfer",
	"20260722020015_create_iam_protected_head_initialize_v1",
	"20260722020016_create_iam_owner_transfer_prepare_v1",
	"20260722020017_create_iam_owner_transfer_ack_v1",
	"20260722020018_create_iam_owner_transfer_commit_v1",
	"20260722020019_create_iam_owner_transfer_compensate_prepare_v1",
	"20260722020020_create_iam_owner_transfer_compensate_apply_v1",
	"20260723040000_create_iam_platform_current_owner_projection_v1",
}

var ownerTransferTenantMigrations = []string{
	"20260722021001_guard_iam_platform_owner_user",
	"20260722021008_create_iam_owner_transfer_observability",
	"20260722021009_create_iam_owner_transfer_prepare_v1",
	"20260722021010_create_iam_owner_transfer_ack_v1",
	"20260722021011_create_iam_owner_transfer_commit_v1",
	"20260722021012_create_iam_owner_transfer_compensate_prepare_v1",
	"20260722021013_create_iam_owner_transfer_compensate_apply_v1",
	"20260722021014_grant_iam_owner_transfer_v1",
}

func prepareOwnerTransferPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase, origin string,
) ownerTransferFixture {
	t.Helper()
	migrator, err := database.migrator.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = migrator.Close() })
	if _, err := migrator.ExecContext(ctx, ownerTransferFoundation); err != nil {
		t.Fatalf("prepare owner transfer foundation: %v", err)
	}
	if _, err := migrator.ExecContext(ctx, "SET search_path TO platform,pg_catalog"); err != nil {
		t.Fatal(err)
	}
	for _, name := range ownerTransferPlatformMigrations {
		applyIAMPlatformGuardDocument(t, ctx, migrator, name,
			readIAMPlatformAuthMigration(t, origin, name))
	}
	seedOwnerTransferAuthority(t, ctx, migrator)
	if _, err := migrator.ExecContext(ctx, "SET search_path TO t_owner,pg_catalog"); err != nil {
		t.Fatal(err)
	}
	for _, name := range ownerTransferTenantMigrations {
		applyIAMPlatformGuardDocument(t, ctx, migrator, name,
			readOwnerTransferTenantMigration(t, origin, name))
	}
	runtime, err := database.runtime.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	if _, err := runtime.ExecContext(ctx, "SET search_path TO t_owner,pg_catalog"); err != nil {
		t.Fatal(err)
	}
	return ownerTransferFixture{migrator: migrator, runtime: runtime}
}

func readOwnerTransferTenantMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", "tenant", origin, name+"."+origin+".sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
