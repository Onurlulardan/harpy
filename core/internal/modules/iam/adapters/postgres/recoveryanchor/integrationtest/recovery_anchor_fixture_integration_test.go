//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const (
	testEnvironmentID = "11111111-1111-4111-8111-111111111111"
	testTenantID      = "22222222-2222-4222-8222-222222222222"
	testOwnerID       = "33333333-3333-4333-8333-333333333333"
	testBindingID     = "44444444-4444-4444-8444-444444444444"
)

func prepareRecoveryAnchorDatabase(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	execRecoveryAnchorSQL(t, ctx, database,
		`CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER`,
		`CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER`,
		`GRANT harpy_runtime TO harpy_v2_runtime`,
		`CREATE ROLE harpy_backup_reader NOLOGIN NOINHERIT NOSUPERUSER`,
		`CREATE ROLE harpy_backup LOGIN NOINHERIT NOSUPERUSER`,
		`GRANT harpy_backup_reader TO harpy_backup`,
		`CREATE ROLE harpy_restore LOGIN NOINHERIT NOSUPERUSER`,
		`CREATE SCHEMA platform`,
	)
}

func prepareRecoveryAnchorPrerequisites(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	execRecoveryAnchorSQL(t, ctx, database,
		`CREATE TABLE platform.environment_identity(
			environment_id uuid PRIMARY KEY,database_system_identifier text NOT NULL)`,
		`CREATE TABLE platform.tenancy_deployment_policy(
			environment_id uuid PRIMARY KEY,deployment_mode text NOT NULL)`,
		`CREATE TABLE platform.tenancy_tenant_bindings(
			environment_id uuid NOT NULL,tenant_id uuid NOT NULL,purpose text NOT NULL)`,
		`CREATE TABLE platform.tenants(
			id uuid PRIMARY KEY,slug text NOT NULL,schema_name name NOT NULL,deleted_at timestamptz)`,
		`CREATE TABLE platform.iam_platform_owner_policies(
			environment_id uuid PRIMARY KEY,owner_binding_id uuid NOT NULL,revision bigint NOT NULL)`,
		`CREATE TABLE platform.iam_protected_authority_heads(
			environment_id uuid PRIMARY KEY,registry_binding_id uuid NOT NULL,generation bigint NOT NULL,
			state_sha256 bytea NOT NULL,ack_sha256 bytea NOT NULL,revision bigint NOT NULL,
			updated_at timestamptz NOT NULL)`,
		`CREATE TABLE platform.iam_platform_authorization_clocks(
			environment_id uuid PRIMARY KEY,authorization_clock bigint NOT NULL,
			revision bigint NOT NULL,updated_at timestamptz NOT NULL)`,
		`CREATE TABLE platform.platformauthority_bindings(
			binding_id uuid PRIMARY KEY,environment_id uuid NOT NULL,deployment_mode text NOT NULL,
			signer_generation bigint NOT NULL,signer_kid text NOT NULL,signer_status text NOT NULL,
			recovery_trust_bundle_sha256 text NOT NULL)`,
		`CREATE TABLE platform.platformauthority_operations(
			operation_id uuid PRIMARY KEY,purpose text NOT NULL,binding_id uuid NOT NULL,
			scope_kind text NOT NULL,scope_id uuid NOT NULL,generation bigint NOT NULL,
			previous_head_sha256 text NOT NULL,state_sha256 text NOT NULL,
			proposal_sha256 text NOT NULL,status text NOT NULL,ack_sha256 text)`,
		`CREATE FUNCTION platform.iam_platform_current_owner_projection_v1(uuid)
			RETURNS TABLE(environment_id uuid,owner_binding_id uuid,owner_policy_revision bigint)
			LANGUAGE sql AS $function$ SELECT environment_id,owner_binding_id,revision
			FROM platform.iam_platform_owner_policies WHERE environment_id=$1 $function$`,
		`CREATE FUNCTION platform.iam_platform_current_owner_v1(uuid,uuid)
			RETURNS boolean LANGUAGE sql AS $function$
			SELECT $1='22222222-2222-4222-8222-222222222222'::uuid
			AND $2='33333333-3333-4333-8333-333333333333'::uuid $function$`,
		`CREATE FUNCTION platform.tenancy_get_primary_tenant_target_v2()
			RETURNS TABLE(environment_id uuid,deployment_mode text,tenant_id uuid,purpose text,
			slug text,schema_name name,tenant_set_sha256 bytea) LANGUAGE sql AS $function$
			SELECT binding.environment_id,'onpremise',binding.tenant_id,binding.purpose,
			tenant.slug,tenant.schema_name,pg_catalog.sha256(pg_catalog.convert_to(
			'tenant-set.v1|'||binding.tenant_id::text||'|'||tenant.schema_name::text,'UTF8'))
			FROM platform.tenancy_tenant_bindings binding JOIN platform.tenants tenant
			ON tenant.id=binding.tenant_id $function$`,
	)
	seedRecoveryAnchorPrerequisites(t, ctx, database)
}

func seedRecoveryAnchorPrerequisites(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	execRecoveryAnchorSQL(t, ctx, database,
		`INSERT INTO platform.environment_identity VALUES(
			'11111111-1111-4111-8111-111111111111','987654321')`,
		`INSERT INTO platform.tenancy_deployment_policy VALUES(
			'11111111-1111-4111-8111-111111111111','onpremise')`,
		`INSERT INTO platform.tenancy_tenant_bindings VALUES(
			'11111111-1111-4111-8111-111111111111',
			'22222222-2222-4222-8222-222222222222','onpremise_primary')`,
		`INSERT INTO platform.tenants VALUES(
			'22222222-2222-4222-8222-222222222222','primary','t_primary',NULL)`,
		`INSERT INTO platform.iam_platform_owner_policies VALUES(
			'11111111-1111-4111-8111-111111111111',
			'33333333-3333-4333-8333-333333333333',1)`,
		`INSERT INTO platform.platformauthority_bindings VALUES(
			'44444444-4444-4444-8444-444444444444',
			'11111111-1111-4111-8111-111111111111','onpremise',1,'kid-1','active',repeat('a',64))`,
		`INSERT INTO platform.iam_protected_authority_heads VALUES(
			'11111111-1111-4111-8111-111111111111',
			'44444444-4444-4444-8444-444444444444',1,
			decode(repeat('11',32),'hex'),decode(repeat('22',32),'hex'),1,clock_timestamp())`,
		`INSERT INTO platform.iam_platform_authorization_clocks VALUES(
			'11111111-1111-4111-8111-111111111111',1,1,clock_timestamp())`,
		`INSERT INTO platform.platformops_installation_recovery_identities(
			operation_id,environment_id,tenant_id,deployment_mode,identity_state,
			signing_generation,trust_bundle_sha256,manifest_sha256,created_at)
			VALUES('10101010-1010-4010-8010-101010101010',
			'11111111-1111-4111-8111-111111111111',
			'22222222-2222-4222-8222-222222222222','onpremise','active',1,
			repeat('3',64),repeat('3',64),clock_timestamp())`,
	)
}

func execRecoveryAnchorSQL(
	t *testing.T, ctx context.Context, database *sql.DB, statements ...string,
) {
	t.Helper()
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("execute recovery-anchor fixture: %v", err)
		}
	}
}
