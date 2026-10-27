//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

const analyticsCutoverEndpoint = "11111111-1111-4111-8111-111111111161"

type analyticsCutoverAuthority struct {
	origin, schema, tenantID string
	operationID, runID       string
	leaseManifest, finalLSN  string
}

func prepareAnalyticsProtectedPG18(t *testing.T, ctx context.Context,
	database *sql.DB, origin string,
) {
	t.Helper()
	prepareAnalyticsPrerequisites(t, ctx, database, origin)
	applyAnalyticsCutoverDependencies(t, ctx, database)
	if origin == "v1_adopted" {
		for _, schema := range []string{"t_analytics_a", "t_analytics_b"} {
			statement := fmt.Sprintf(`CREATE TABLE %s.openapi_endpoints(
id uuid PRIMARY KEY,tenant_id uuid NOT NULL,slug text NOT NULL,name text NOT NULL,
description text NOT NULL,method text NOT NULL,config_json jsonb NOT NULL,
is_active boolean NOT NULL,is_public boolean NOT NULL,created_by uuid NOT NULL,
updated_by uuid NOT NULL,created_at timestamptz NOT NULL,
updated_at timestamptz NOT NULL,deleted_at timestamptz)`, schema)
			if _, err := database.ExecContext(ctx, statement); err != nil {
				t.Fatal(err)
			}
		}
		seedAnalyticsCutoverSource(t, ctx, database)
	}
	applyAnalyticsDocuments(t, ctx, database, "", analyticsDocuments(t, origin, "platform"))
	reconciled, remaining := reconcileAnalyticsAuthorities(t, ctx, database, true)
	if reconciled != 2 || remaining != 0 {
		t.Fatalf("Analytics authority reconcile = %d/%d, want 2/0",
			reconciled, remaining)
	}
	for _, schema := range []string{"t_analytics_a", "t_analytics_b"} {
		applyAnalyticsDocuments(t, ctx, database, schema,
			analyticsDocuments(t, origin, "tenant"))
	}
}

func seedAnalyticsCutoverSource(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO t_analytics_a.openapi_endpoints(
id,tenant_id,slug,name,description,method,config_json,is_active,is_public,
created_by,updated_by,created_at,updated_at,deleted_at) VALUES(
$1,$2,'legacy-report','Legacy report','initial','GET',
'{"path":"/legacy/report"}'::jsonb,true,true,$3,$3,
'2026-07-21T10:00:00.123Z','2026-07-21T10:00:00.123Z',NULL)`,
		analyticsCutoverEndpoint, analyticsTenantA, analyticsCutoverRequester)
	if err != nil {
		t.Fatal(err)
	}
}

func prepareAnalyticsCutoverAuthority(t *testing.T, ctx context.Context,
	database *sql.DB, origin string,
) analyticsCutoverAuthority {
	t.Helper()
	authority := analyticsCutoverAuthority{origin: origin, schema: "t_analytics_a",
		tenantID: analyticsTenantA, operationID: "018f2000-0000-7000-8000-000000000161",
		runID:         "018f2000-0000-7000-8000-000000000162",
		leaseManifest: fmt.Sprintf("%064x", 161), finalLSN: "0/70"}
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
FROM pg_catalog.pg_roles WHERE rolname='harpy_adopter'`).Scan(&leaseExpiry); err != nil {
		t.Fatal(err)
	}
	execAnalyticsCutoverSeed(t, ctx, database, `UPDATE platform.cutover_groups
SET generation=2,state='shadow',writer='v1' WHERE scope_kind='tenant'
AND scope_id=$1 AND group_id='analytics'`, authority.tenantID)
	execAnalyticsCutoverSeed(t, ctx, database, `UPDATE platform.domain_cutovers
SET generation=2,state='shadow',writer='v1' WHERE scope_kind='tenant'
AND scope_id=$1 AND domain='analytics'`, authority.tenantID)
	execAnalyticsCutoverSeed(t, ctx, database, `INSERT INTO platform.domain_cutovers
(scope_kind,scope_id,domain,generation,state,writer) VALUES
('tenant',$1,'iam',1,'v2','v2')`, authority.tenantID)
	seedAnalyticsCutoverOperation(t, ctx, database, authority, leaseExpiry)
	return authority
}

func seedAnalyticsCutoverOperation(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority, leaseExpiry time.Time,
) {
	t.Helper()
	source := "openapi_endpoints"
	if authority.origin == "fresh" {
		source = "analytics_saved_queries"
	}
	execAnalyticsCutoverSeed(t, ctx, database, `INSERT INTO platform.cutover_operations(
operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
group_id,domain,cutover_generation,lease_generation,lease_owner,lease_expires_at,
last_heartbeat_at,lease_manifest_sha256,plan_sha256,transform_manifest_sha256,
source_system_identifier,source_database_oid,target_system_identifier,
target_database_oid,publication_name,slot_name,state,created_at,updated_at)
SELECT $2,$3,'018f2000-0000-7000-8000-000000000163',
'018f2000-0000-7000-8000-000000000164',
'018f2000-0000-7000-8000-000000000165','tenant',$1,'analytics','analytics',2,1,
$4,$5,clock_timestamp(),$6,$7,$8,control.system_identifier,database.oid,
control.system_identifier,database.oid,'analytics_phase8_pub','analytics_phase8_slot',
'shadow',clock_timestamp(),clock_timestamp()
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database ON database.datname=current_database()`,
		authority.tenantID, authority.operationID, authority.runID,
		analyticsCutoverRequester, leaseExpiry, authority.leaseManifest,
		fmt.Sprintf("%064x", 162), fmt.Sprintf("%064x", 163))
	seedAnalyticsCutoverBindings(t, ctx, database, authority, source)
}

func seedAnalyticsCutoverBindings(t *testing.T, ctx context.Context,
	database *sql.DB, authority analyticsCutoverAuthority, source string,
) {
	t.Helper()
	procedure := authority.schema + ".analytics_v1_adoption_apply_v1" +
		"(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)"
	execAnalyticsCutoverSeed(t, ctx, database, `INSERT INTO platform.cutover_operation_relations(
operation_id,run_id,ordinal,source_schema_oid,source_relation_oid,
source_schema_name,source_relation_name,target_schema_oid,target_relation_oid,
target_schema_name,target_relation_name,replica_identity)
SELECT $1,$2,1,namespace.oid,source.oid,$3::text,$4::text,
namespace.oid,target.oid,$3::text,
'analytics_v1_adoption_stage','primary_key' FROM pg_catalog.pg_namespace AS namespace
JOIN pg_catalog.pg_class AS source ON source.relnamespace=namespace.oid
 AND source.relname=$4::text
JOIN pg_catalog.pg_class AS target ON target.relnamespace=namespace.oid
 AND target.relname='analytics_v1_adoption_stage' WHERE namespace.nspname=$3::text`,
		authority.operationID, authority.runID, authority.schema, source)
	execAnalyticsCutoverSeed(t, ctx, database, `INSERT INTO
platform.cutover_operation_checkpoint_keys(operation_id,run_id,relation_ordinal,
ordinal,checkpoint_key) VALUES($1,$2,1,1,'openapi_endpoints')`,
		authority.operationID, authority.runID)
	execAnalyticsCutoverSeed(t, ctx, database, `INSERT INTO
platform.cutover_operation_procedures(operation_id,run_id,ordinal,
relation_ordinal,procedure_kind,procedure_oid,procedure_schema,procedure_name,
identity_arguments,procedure_signature,procedure_contract_id)
VALUES($1,$2,1,1,'apply',$4::regprocedure::oid,$3,
'analytics_v1_adoption_apply_v1',
'uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text',
$4,'analytics.v1_adoption.v1')`, authority.operationID, authority.runID,
		authority.schema, procedure)
}

func execAnalyticsCutoverSeed(t *testing.T, ctx context.Context,
	database *sql.DB, statement string, arguments ...any,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, statement, arguments...); err != nil {
		t.Fatalf("seed Analytics cutover: %v; SQL: %s", err, statement)
	}
}
