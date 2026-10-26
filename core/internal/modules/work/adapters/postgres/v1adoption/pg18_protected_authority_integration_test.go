//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

type protectedWorkAuthority struct {
	schema        string
	tenantID      string
	operationID   string
	runID         string
	leaseManifest string
	finalLSN      string
}

func prepareProtectedWorkAuthority(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) protectedWorkAuthority {
	t.Helper()
	authority := protectedWorkAuthority{schema: schema, tenantID: workTenantID(schema),
		operationID: workUUID(41, 1).String(), runID: workUUID(41, 2).String(),
		leaseManifest: fmt.Sprintf("%064x", 41), finalLSN: "0/90"}
	wantedExpiry := time.Now().UTC().Add(20 * time.Minute).Truncate(time.Microsecond)
	expiryLiteral := wantedExpiry.Format(time.RFC3339Nano)
	sealRoles := fmt.Sprintf(`
ALTER ROLE harpy_adopter NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0
VALID UNTIL '%s' PASSWORD NULL;
ALTER ROLE harpy_cutover_decoder NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0
VALID UNTIL '%s' PASSWORD NULL;`, expiryLiteral, expiryLiteral)
	if _, err := database.ExecContext(ctx, sealRoles); err != nil {
		t.Fatalf("seal Work cutover roles: %v", err)
	}
	var leaseExpiry time.Time
	if err := database.QueryRowContext(ctx, `SELECT rolvaliduntil
FROM pg_catalog.pg_roles WHERE rolname='harpy_adopter'`).Scan(&leaseExpiry); err != nil {
		t.Fatal(err)
	}
	seedProtectedWorkOperation(t, ctx, database, authority, leaseExpiry)
	if _, err := database.ExecContext(ctx, "GRANT USAGE ON SCHEMA "+schema+
		" TO harpy_adopter"); err != nil {
		t.Fatal(err)
	}
	return authority
}

func seedProtectedWorkOperation(t *testing.T, ctx context.Context, database *sql.DB,
	a protectedWorkAuthority, leaseExpiry time.Time,
) {
	t.Helper()
	const requestedBy = "41000000-0000-4000-8000-000000000001"
	const approvedBy = "41000000-0000-4000-8000-000000000002"
	execProtectedWorkSQL(t, ctx, database, `INSERT INTO platform.domain_cutovers
(scope_kind,scope_id,domain,state,writer,generation,requested_by,approved_by,
 reason,change_reference)
VALUES ('tenant',$1,'work','quiescing','none',9,$2,$3,'PG18 Work cutover','phase10')`,
		a.tenantID, requestedBy, approvedBy)
	execProtectedWorkSQL(t, ctx, database, `INSERT INTO platform.cutover_groups
(scope_kind,scope_id,group_id,state,writer,generation,requested_by,approved_by,
 reason,change_reference)
VALUES ('tenant',$1,'work','quiescing','none',9,$2,$3,'PG18 Work cutover','phase10')`,
		a.tenantID, requestedBy, approvedBy)
	execProtectedWorkSQL(t, ctx, database, `INSERT INTO platform.cutover_group_members
(scope_kind,scope_id,group_id,ordinal,domain) VALUES ('tenant',$1,'work',1,'work')`,
		a.tenantID)
	execProtectedWorkSQL(t, ctx, database, `INSERT INTO platform.cutover_operations
(operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
 group_id,domain,cutover_generation,lease_generation,lease_owner,lease_expires_at,
 last_heartbeat_at,lease_manifest_sha256,plan_sha256,transform_manifest_sha256,
 source_system_identifier,source_database_oid,target_system_identifier,
 target_database_oid,publication_name,slot_name,state,created_at,updated_at)
SELECT $3,$4,'41000000-0000-4000-8000-000000000010',
 '41000000-0000-4000-8000-000000000011',
 '41000000-0000-4000-8000-000000000012','tenant',$1,'work','work',9,1,$2,$5,
 pg_catalog.clock_timestamp(),$6,$7,$8,control.system_identifier,database.oid,
 control.system_identifier,database.oid,'work_phase10_pub','work_phase10_slot',
 'quiescing',pg_catalog.clock_timestamp(),pg_catalog.clock_timestamp()
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database ON database.datname=current_database()`,
		a.tenantID, requestedBy, a.operationID, a.runID, leaseExpiry,
		a.leaseManifest, fmt.Sprintf("%064x", 42), fmt.Sprintf("%064x", 43))
	seedProtectedWorkBindings(t, ctx, database, a)
}

func seedProtectedWorkBindings(t *testing.T, ctx context.Context, database *sql.DB,
	a protectedWorkAuthority,
) {
	t.Helper()
	procedure := a.schema + ".work_v1_adoption_apply_v1" +
		"(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)"
	execProtectedWorkSQL(t, ctx, database, `INSERT INTO platform.cutover_operation_relations
(operation_id,run_id,ordinal,source_schema_oid,source_relation_oid,
 source_schema_name,source_relation_name,target_schema_oid,target_relation_oid,
 target_schema_name,target_relation_name,replica_identity)
SELECT $1,$2,1,namespace.oid,source.oid,$3::text,'task_boards',namespace.oid,target.oid,
 $3::text,'work_v1_adoption_stage','primary_key'
FROM pg_catalog.pg_namespace AS namespace
JOIN pg_catalog.pg_class AS source ON source.relnamespace=namespace.oid
 AND source.relname='task_boards'
JOIN pg_catalog.pg_class AS target ON target.relnamespace=namespace.oid
 AND target.relname='work_v1_adoption_stage'
WHERE namespace.nspname=$3::text`, a.operationID, a.runID, a.schema)
	execProtectedWorkSQL(t, ctx, database, `INSERT INTO platform.cutover_operation_checkpoint_keys
(operation_id,run_id,relation_ordinal,ordinal,checkpoint_key)
SELECT $1,$2,1,key.ordinal,key.name FROM (VALUES
 (1,'task_boards'),(2,'task_board_columns'),(3,'board_members'),
 (4,'board_tasks'),(5,'board_task_comments'),(6,'calendar_events')
) AS key(ordinal,name)`, a.operationID, a.runID)
	execProtectedWorkSQL(t, ctx, database, `INSERT INTO platform.cutover_operation_procedures
(operation_id,run_id,ordinal,relation_ordinal,procedure_kind,procedure_oid,
 procedure_schema,procedure_name,identity_arguments,procedure_signature,
 procedure_contract_id)
VALUES ($1,$2,1,1,'apply',$4::regprocedure::oid,$3,
 'work_v1_adoption_apply_v1',
 'uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text',
 $4,'work.v1_adoption.v1')`, a.operationID, a.runID, a.schema, procedure)
	execProtectedWorkSQL(t, ctx, database, `INSERT INTO platform.cutover_change_journal
(operation_id,run_id,scope_kind,scope_id,domain,cutover_generation,source_xid,
 source_session_user)
VALUES ($1,$2,'tenant',$3,'work',9,'4101'::xid8,'harpy_v1_runtime')`,
		a.operationID, a.runID, a.tenantID)
}

func execProtectedWorkSQL(t *testing.T, ctx context.Context, database *sql.DB,
	statement string, arguments ...any,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, statement, arguments...)
	if err != nil {
		t.Fatalf("seed protected Work authority: %v; SQL: %s", err, statement)
	}
}
