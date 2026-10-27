//go:build integration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"
)

func setupRecoveryIdentity(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	statements := []string{
		`CREATE TABLE platform.environment_identity(
			environment_id uuid PRIMARY KEY,database_system_identifier text NOT NULL)`,
		`CREATE TABLE platform.tenancy_tenant_bindings(
			environment_id uuid NOT NULL,tenant_id uuid NOT NULL,purpose text NOT NULL)`,
		`CREATE TABLE platform.tenancy_deployment_policy(
			singleton boolean PRIMARY KEY CHECK(singleton),
			environment_id uuid NOT NULL UNIQUE,
			deployment_mode text NOT NULL CHECK(deployment_mode IN ('onpremise','saas')),
			maximum_customer_tenants bigint,created_at timestamptz NOT NULL)`,
		`CREATE TABLE platform.tenants(
			id uuid PRIMARY KEY,slug text NOT NULL,schema_name name NOT NULL,
			deleted_at timestamptz)`,
		`CREATE TABLE platform.iam_platform_owner_policies(
			environment_id uuid PRIMARY KEY,owner_binding_id uuid NOT NULL,revision bigint NOT NULL)`,
		`CREATE FUNCTION platform.iam_platform_current_owner_projection_v1(uuid)
			RETURNS TABLE(environment_id uuid,owner_binding_id uuid,
			owner_policy_revision bigint) LANGUAGE sql AS $function$
			SELECT policy.environment_id,policy.owner_binding_id,policy.revision
			FROM platform.iam_platform_owner_policies AS policy
			WHERE policy.environment_id=$1
			$function$`,
		`CREATE FUNCTION platform.iam_platform_current_owner_v1(uuid,uuid)
			RETURNS boolean LANGUAGE sql AS $function$
			SELECT $1='22222222-2222-4222-8222-222222222222'::uuid
			AND $2='33333333-3333-4333-8333-333333333333'::uuid
			$function$`,
		`CREATE FUNCTION platform.tenancy_get_primary_tenant_target_v2()
			RETURNS TABLE(environment_id uuid,deployment_mode text,tenant_id uuid,
			purpose text,slug text,schema_name name,tenant_set_sha256 bytea)
			LANGUAGE sql AS $function$
			SELECT binding.environment_id,'onpremise'::text,binding.tenant_id,
			binding.purpose,tenant.slug,tenant.schema_name,
			pg_catalog.sha256(pg_catalog.convert_to(
			'tenant-set.v1|' || binding.tenant_id::text || '|' ||
			tenant.schema_name::text,'UTF8'))
			FROM platform.tenancy_tenant_bindings AS binding
			JOIN platform.tenants AS tenant ON tenant.id=binding.tenant_id
			$function$`,
		`INSERT INTO platform.environment_identity VALUES(
			'11111111-1111-4111-8111-111111111111','987654321')`,
		`INSERT INTO platform.tenancy_tenant_bindings VALUES(
			'11111111-1111-4111-8111-111111111111',
			'22222222-2222-4222-8222-222222222222','onpremise_primary')`,
		`INSERT INTO platform.tenancy_deployment_policy VALUES(
			true,'11111111-1111-4111-8111-111111111111',
			'onpremise',0,statement_timestamp())`,
		`INSERT INTO platform.tenants VALUES(
			'22222222-2222-4222-8222-222222222222','primary','t_primary',NULL)`,
		`INSERT INTO platform.iam_platform_owner_policies VALUES(
			'11111111-1111-4111-8111-111111111111',
			'33333333-3333-4333-8333-333333333333',1)`,
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Date(2026, 7, 23, 11, 0, 0, 0, time.UTC)
	_, err := database.ExecContext(ctx, `INSERT INTO
		platform.platformops_installation_recovery_identities(
		operation_id,environment_id,tenant_id,deployment_mode,identity_state,
		signing_generation,trust_bundle_sha256,manifest_sha256,created_at)
		VALUES($1,$2,$3,'onpremise','active',1,$4,$4,$5)`,
		"10101010-1010-4010-8010-101010101010", environmentID, primaryID, hashA, now)
	if err != nil {
		t.Fatal(err)
	}
	assertCurrentRecoveryIdentity(t, ctx, database)
}

func assertCurrentRecoveryIdentity(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var environment, mode, primary, tenantDigest, owner, lineage, system string
	var ownerGeneration int64
	err := database.QueryRowContext(ctx,
		`SELECT * FROM platform.platformops_current_environment_identity_v1($1)`,
		environmentID).Scan(&environment, &mode, &primary, &tenantDigest, &owner,
		&ownerGeneration, &lineage, &system)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("tenant-set.v1|" + primaryID + "|t_primary"))
	if environment != environmentID || mode != "onpremise" || primary != primaryID ||
		tenantDigest != hex.EncodeToString(digest[:]) || owner != ownerID ||
		ownerGeneration != 1 || lineage != environmentID || system == "" {
		t.Fatalf("unexpected current recovery identity")
	}
}
