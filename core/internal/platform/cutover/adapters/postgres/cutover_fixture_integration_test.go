//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	integrationOperationID = "10000000-0000-4000-8000-000000000001"
	integrationRunID       = "10000000-0000-4000-8000-000000000002"
	integrationLeaseID     = "10000000-0000-4000-8000-000000000003"
	integrationReleaseID   = "10000000-0000-4000-8000-000000000004"
	integrationEnvironment = "10000000-0000-4000-8000-000000000005"
	integrationScopeID     = "10000000-0000-4000-8000-000000000006"
	integrationOwnerID     = "10000000-0000-4000-8000-000000000007"
	integrationTenantID    = "10000000-0000-4000-8000-000000000008"
	integrationReleaseRun  = "10000000-0000-4000-8000-000000000009"
	integrationTenantBID   = "10000000-0000-4000-8000-000000000010"
	integrationPriorRun    = "10000000-0000-4000-8000-000000000011"
)

func prepareReadinessSurfaces(t *testing.T, ctx context.Context, admin *pgx.Conn, origin string) {
	t.Helper()
	prepareCutoverEnvironmentIdentity(t, ctx, admin, origin)
	_, err := admin.Exec(ctx, `
CREATE TABLE platform.release_generations (
    migration_run_id uuid PRIMARY KEY, release_id uuid NOT NULL, environment_id uuid NOT NULL, database_origin text NOT NULL,
    database_generation bigint NOT NULL, minimum_runtime_generation bigint NOT NULL,
    migration_manifest_sha256 text NOT NULL, extension_manifest_sha256 text NOT NULL
);
CREATE TABLE platform.schema_contracts (
    component text PRIMARY KEY, verified_run_id uuid NOT NULL,
    database_generation bigint NOT NULL, minimum_runtime_generation bigint NOT NULL,
    migration_manifest_sha256 text NOT NULL, extension_manifest_sha256 text NOT NULL,
    status text NOT NULL
);
CREATE TABLE platform.tenants (
    id uuid PRIMARY KEY,
    schema_name text NOT NULL UNIQUE,
    deleted_at timestamptz
);
CREATE TABLE platform.integration_business_rows (id uuid PRIMARY KEY);
CREATE SCHEMA legacy_source;
CREATE TABLE legacy_source.allowed_rows (id uuid PRIMARY KEY);
INSERT INTO legacy_source.allowed_rows VALUES ('80000000-0000-4000-8000-000000000001');
CREATE TABLE platform.tenant_migration_checkpoints (
    tenant_id uuid NOT NULL, owner text NOT NULL, logical_version bigint NOT NULL,
    database_origin text NOT NULL, selected_path text NOT NULL,
    selected_sha256 text NOT NULL, run_id uuid NOT NULL, state text NOT NULL,
    PRIMARY KEY (tenant_id, owner, logical_version)
);
CREATE FUNCTION platform.platform_migrator_reject_mutation()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$ BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'immutable integration fixture';
END; $function$;
GRANT USAGE ON SCHEMA platform TO harpy_adopter, harpy_cutover_finalizer;
GRANT USAGE ON SCHEMA legacy_source TO harpy_cutover_decoder;
GRANT SELECT ON legacy_source.allowed_rows TO harpy_cutover_decoder;`)
	if err != nil {
		t.Fatalf("prepare cutover readiness surfaces: %v", err)
	}
	prepareSuccessorModuleStubs(t, ctx, admin)
	prepareCutoverOwnerPrerequisiteSurfaces(t, ctx, admin)
	applyTenantRequirementAuthorityMigrations(t, ctx, admin)
}

func applyTenantRequirementAuthorityMigrations(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	directory := filepath.Join("..", "..", "..", "migrator", "migrations", "fresh")
	for _, name := range []string{
		"20260714012910_create_tenant_requirement_catalogs.fresh.sql",
		"20260714012920_create_tenant_requirements.fresh.sql",
		"20260714012930_create_tenant_requirement_verifier_v2.fresh.sql",
	} {
		document, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := connection.Exec(ctx, string(document)); err != nil {
			t.Fatalf("apply tenant requirement authority migration %s: %v", name, err)
		}
	}
}

func seedCutoverOperation(
	t *testing.T, ctx context.Context, admin *pgx.Conn, expiresAt time.Time,
) {
	t.Helper()
	digestA, digestB, digestC := strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64)
	_, err := admin.Exec(ctx, `
INSERT INTO platform.domain_cutovers (
    scope_kind, scope_id, domain, state, writer, generation, requested_by,
    reason, change_reference
) VALUES ('tenant', $1, 'iam', 'shadow', 'v1', 1, $2, 'integration', 'CUTOVER-1');
INSERT INTO platform.domain_cutovers (
    scope_kind, scope_id, domain, state, writer, generation, requested_by,
    reason, change_reference
) VALUES ('tenant', $1, 'system_resource_layout', 'shadow', 'v1', 1, $2, 'integration', 'CUTOVER-1');
INSERT INTO platform.cutover_groups (
    scope_kind, scope_id, group_id, state, writer, generation, requested_by,
    reason, change_reference
) VALUES ('tenant', $1, 'iam_system_resources', 'shadow', 'v1', 1, $2, 'integration', 'CUTOVER-1');
INSERT INTO platform.cutover_group_members
    (scope_kind, scope_id, group_id, ordinal, domain)
VALUES ('tenant', $1, 'iam_system_resources', 1, 'iam'),
    ('tenant', $1, 'iam_system_resources', 2, 'system_resource_layout');
INSERT INTO platform.cutover_operations (
    operation_id, run_id, lease_id, release_id, environment_id, scope_kind,
    scope_id, group_id, domain, cutover_generation, lease_generation, lease_owner,
    lease_expires_at, last_heartbeat_at, lease_manifest_sha256, plan_sha256,
    transform_manifest_sha256, source_system_identifier, source_database_oid,
    target_system_identifier, target_database_oid, publication_name, slot_name, state
) SELECT $3, $4, $5, $6, $7, 'tenant', $1, 'iam_system_resources', 'iam', 1, 1, $2,
    $11, clock_timestamp(), $8, $9, $10,
    control.system_identifier, database_row.oid, control.system_identifier,
    database_row.oid, 'cutover_iam', 'cutover_iam_slot', 'copying'
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database_row ON database_row.datname = current_database();
INSERT INTO platform.cutover_operation_relations (
    operation_id, run_id, ordinal, source_schema_oid, source_relation_oid,
    source_schema_name, source_relation_name, target_schema_oid, target_relation_oid,
    target_schema_name, target_relation_name, replica_identity
) VALUES ($3, $4, 1, 2200, 17000, 'legacy', 'iam_source',
    2200, 18000, 'platform', 'iam_target', 'primary_key');
INSERT INTO platform.cutover_operation_checkpoint_keys (
    operation_id, run_id, relation_ordinal, ordinal, checkpoint_key
) VALUES ($3, $4, 1, 1, 'main');
INSERT INTO platform.cutover_operation_columns (
    operation_id, run_id, relation_ordinal, ordinal, source_attnum,
    target_attnum, transform_id, primary_key
) VALUES ($3, $4, 1, 1, 1, 1, 'identity.uuid', true);
INSERT INTO platform.cutover_operation_procedures (
    operation_id, run_id, ordinal, relation_ordinal, procedure_kind,
    procedure_oid, procedure_schema, procedure_name, identity_arguments,
    procedure_signature, procedure_contract_id
) VALUES ($3, $4, 1, 1, 'apply',
    'platform.cutover_probe_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)'::regprocedure::oid,
    'platform', 'cutover_probe_apply_v1',
    'uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text',
    'platform.cutover_probe_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)',
	'platform.cutover_probe.v1');`, pgx.QueryExecModeSimpleProtocol,
		integrationScopeID, integrationOwnerID, integrationOperationID, integrationRunID,
		integrationLeaseID, integrationReleaseID, integrationEnvironment,
		digestA, digestB, digestC,
		expiresAt,
	)
	if err != nil {
		t.Fatalf("seed cutover operation: %v", err)
	}
}

func setCutoverRoleLease(t *testing.T, ctx context.Context, admin *pgx.Conn, phase string) {
	t.Helper()
	statements := map[string]string{
		"adoption_seal": `ALTER ROLE harpy_adopter NOLOGIN NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL;
ALTER ROLE harpy_cutover_decoder NOLOGIN NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL;`,
		"adoption_cleanup": `ALTER ROLE harpy_adopter NOLOGIN NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL VALID UNTIL 'infinity';
ALTER ROLE harpy_cutover_decoder NOLOGIN NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL VALID UNTIL 'infinity';`,
		"finalizer": `ALTER ROLE harpy_cutover_finalizer LOGIN NOINHERIT NOSUPERUSER
NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1 PASSWORD 'integration-leased-secret';`,
		"finalizer_cleanup":     `ALTER ROLE harpy_cutover_finalizer NOLOGIN NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL VALID UNTIL 'infinity';`,
		"cluster_admin_cleanup": `ALTER ROLE harpy_cluster_admin NOLOGIN NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL VALID UNTIL 'infinity';`,
	}
	statement, found := statements[phase]
	if !found {
		t.Fatalf("unknown cutover role phase %s", phase)
	}
	if _, err := admin.Exec(ctx, statement); err != nil {
		t.Fatalf("set cutover role phase %s: %v", phase, err)
	}
}

func assertCutoverRole(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	role string, login bool, replication bool, connectionLimit int,
) {
	t.Helper()
	var actualLogin, actualReplication, passwordNull bool
	var actualLimit int
	err := admin.QueryRow(ctx, `SELECT role.rolcanlogin, role.rolreplication,
    role.rolconnlimit, role.rolpassword IS NULL FROM pg_catalog.pg_authid AS role
    WHERE role.rolname = $1`, role).Scan(
		&actualLogin, &actualReplication, &actualLimit, &passwordNull,
	)
	if err != nil || actualLogin != login || actualReplication != replication ||
		actualLimit != connectionLimit || (login && passwordNull) || (!login && !passwordNull) {
		t.Fatalf("role %s catalog = %t %t %d %t, err=%v",
			role, actualLogin, actualReplication, actualLimit, passwordNull, err)
	}
	assertCutoverRoleCatalogBoundary(t, ctx, admin, role)
}
