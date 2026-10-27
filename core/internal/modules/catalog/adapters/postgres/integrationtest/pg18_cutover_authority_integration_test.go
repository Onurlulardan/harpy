//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

type catalogCutoverAuthority struct {
	origin, schema, tenantID string
	operationID, runID       string
	leaseManifest, finalLSN  string
}

func prepareCatalogCutoverAuthority(t *testing.T, ctx context.Context,
	database *sql.DB, origin string,
) catalogCutoverAuthority {
	t.Helper()
	authority := catalogCutoverAuthority{origin: origin, schema: "catalog_a",
		tenantID:      catalogTenantA,
		operationID:   "018f2000-0000-7000-8000-000000000061",
		runID:         "018f2000-0000-7000-8000-000000000062",
		leaseManifest: fmt.Sprintf("%064x", 61), finalLSN: "0/90"}
	expires := time.Now().UTC().Add(20 * time.Minute).Truncate(time.Microsecond)
	literal := expires.Format(time.RFC3339Nano)
	statement := fmt.Sprintf(`ALTER ROLE harpy_adopter NOLOGIN NOINHERIT
NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS
CONNECTION LIMIT 0 VALID UNTIL '%s' PASSWORD NULL;
ALTER ROLE harpy_cutover_decoder NOLOGIN NOINHERIT NOSUPERUSER
NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0
VALID UNTIL '%s' PASSWORD NULL;`, literal, literal)
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatal(err)
	}
	var leaseExpiry time.Time
	if err := database.QueryRowContext(ctx, `SELECT rolvaliduntil
FROM pg_catalog.pg_roles WHERE rolname='harpy_adopter'`).Scan(
		&leaseExpiry,
	); err != nil {
		t.Fatal(err)
	}
	execCatalogCutoverSeed(t, ctx, database, `UPDATE platform.cutover_groups
SET state='quiescing',writer='none' WHERE scope_kind='tenant' AND scope_id=$1
AND group_id='catalog'`, authority.tenantID)
	execCatalogCutoverSeed(t, ctx, database, `UPDATE platform.domain_cutovers
SET state='quiescing',writer='none' WHERE scope_kind='tenant' AND scope_id=$1
AND domain='catalog'`, authority.tenantID)
	seedCatalogCutoverOperation(t, ctx, database, authority, leaseExpiry)
	return authority
}

func seedCatalogCutoverOperation(t *testing.T, ctx context.Context,
	database *sql.DB, authority catalogCutoverAuthority, leaseExpiry time.Time,
) {
	t.Helper()
	const actor = "018f2000-0000-7000-8000-000000000010"
	execCatalogCutoverSeed(t, ctx, database, `INSERT INTO platform.cutover_operations(
operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
group_id,domain,cutover_generation,lease_generation,lease_owner,lease_expires_at,
last_heartbeat_at,lease_manifest_sha256,plan_sha256,transform_manifest_sha256,
source_system_identifier,source_database_oid,target_system_identifier,
target_database_oid,publication_name,slot_name,state,created_at,updated_at)
SELECT $2,$3,'018f2000-0000-7000-8000-000000000063',
'018f2000-0000-7000-8000-000000000064',
'018f2000-0000-7000-8000-000000000065','tenant',$1,'catalog','catalog',7,1,
$4,$5,clock_timestamp(),$6,$7,$8,control.system_identifier,database.oid,
control.system_identifier,database.oid,'catalog_phase9_pub','catalog_phase9_slot',
'quiescing',clock_timestamp(),clock_timestamp()
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database ON database.datname=current_database()`,
		authority.tenantID, authority.operationID, authority.runID, actor, leaseExpiry,
		authority.leaseManifest, fmt.Sprintf("%064x", 62), fmt.Sprintf("%064x", 63))
	seedCatalogCutoverBindings(t, ctx, database, authority)
}

func seedCatalogCutoverBindings(t *testing.T, ctx context.Context,
	database *sql.DB, authority catalogCutoverAuthority,
) {
	t.Helper()
	source := "modules"
	if authority.origin == "fresh" {
		source = "catalog_installations"
	}
	procedure := authority.schema + ".catalog_v1_adoption_apply_v1" +
		"(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)"
	execCatalogCutoverSeed(t, ctx, database, `INSERT INTO platform.cutover_operation_relations(
operation_id,run_id,ordinal,source_schema_oid,source_relation_oid,
source_schema_name,source_relation_name,target_schema_oid,target_relation_oid,
target_schema_name,target_relation_name,replica_identity)
SELECT $1,$2,1,namespace.oid,source.oid,$3::text,$4::text,
namespace.oid,target.oid,$3::text,
'catalog_v1_adoption_stage','primary_key'
FROM pg_catalog.pg_namespace AS namespace
JOIN pg_catalog.pg_class AS source ON source.relnamespace=namespace.oid
 AND source.relname=$4::text
JOIN pg_catalog.pg_class AS target ON target.relnamespace=namespace.oid
 AND target.relname='catalog_v1_adoption_stage'
WHERE namespace.nspname=$3::text`, authority.operationID, authority.runID,
		authority.schema, source)
	execCatalogCutoverSeed(t, ctx, database, `INSERT INTO
platform.cutover_operation_checkpoint_keys(
operation_id,run_id,relation_ordinal,ordinal,checkpoint_key)
SELECT $1,$2,1,key.ordinal,key.name FROM (VALUES (1,'modules'),
(2,'module_dependencies'),(3,'module_forms'),(4,'module_schema_versions'),
(5,'module_migration_runs'),(6,'menu_configs')) AS key(ordinal,name)`,
		authority.operationID, authority.runID)
	execCatalogCutoverSeed(t, ctx, database, `INSERT INTO
platform.cutover_operation_procedures(operation_id,run_id,ordinal,
relation_ordinal,procedure_kind,procedure_oid,procedure_schema,procedure_name,
identity_arguments,procedure_signature,procedure_contract_id)
VALUES($1,$2,1,1,'apply',$4::regprocedure::oid,$3,
'catalog_v1_adoption_apply_v1',
'uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text',
$4,'catalog.v1_adoption.v1')`, authority.operationID, authority.runID,
		authority.schema, procedure)
}

func execCatalogCutoverSeed(t *testing.T, ctx context.Context,
	database *sql.DB, statement string, arguments ...any,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, statement, arguments...); err != nil {
		t.Fatalf("seed Catalog cutover authority: %v; SQL: %s", err, statement)
	}
}
