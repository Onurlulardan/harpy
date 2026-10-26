//go:build integration

package postgres

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

const (
	builtCutoverPassword        = "built-cutover-leased-secret"
	builtCutoverExpiredPassword = "built-cutover-expired-secret"
	builtCutoverProcedure       = "platform.cutover_probe_apply_v1(uuid,uuid,text,uuid," +
		"smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)"
)

func readBuiltCutoverDatabaseIdentity(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) (cutoverv1.DatabaseIdentity, uint32) {
	t.Helper()
	var identity cutoverv1.DatabaseIdentity
	var procedureOID uint32
	err := database.admin.QueryRow(ctx, `SELECT current_database(),
    control.system_identifier::text, database_row.oid,
    $1::regprocedure::oid
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database_row ON database_row.datname=current_database()`,
		builtCutoverProcedure).Scan(
		&identity.DatabaseName, &identity.SystemIdentifier, &identity.DatabaseOID, &procedureOID,
	)
	if err != nil {
		t.Fatalf("read built cutover database identity: %v", err)
	}
	identity.DatabaseOrigin = "fresh"
	return identity, procedureOID
}

func seedBuiltCutoverOperation(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	artifacts builtCutoverLeaseArtifacts,
) {
	t.Helper()
	relation := artifacts.lease.Relations[0]
	_, err := database.admin.Exec(ctx, `
INSERT INTO platform.domain_cutovers (
    scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference
) VALUES ('tenant',$1,'iam','shadow','v1',1,$2,'built gate','CUTOVER-BUILT')
ON CONFLICT DO NOTHING;
INSERT INTO platform.domain_cutovers (
    scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference
) VALUES ('tenant',$1,'system_resource_layout','shadow','v1',1,$2,'built gate','CUTOVER-BUILT')
ON CONFLICT DO NOTHING;
INSERT INTO platform.cutover_groups (
    scope_kind,scope_id,group_id,state,writer,generation,requested_by,reason,change_reference
) VALUES ('tenant',$1,'iam_system_resources','shadow','v1',1,$2,'built gate','CUTOVER-BUILT')
ON CONFLICT DO NOTHING;
INSERT INTO platform.cutover_group_members
    (scope_kind,scope_id,group_id,ordinal,domain)
VALUES ('tenant',$1,'iam_system_resources',1,'iam'),
    ('tenant',$1,'iam_system_resources',2,'system_resource_layout')
ON CONFLICT DO NOTHING;
INSERT INTO platform.cutover_operations (
    operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
    group_id,domain,cutover_generation,lease_generation,lease_owner,lease_expires_at,
    last_heartbeat_at,lease_manifest_sha256,plan_sha256,transform_manifest_sha256,
    source_system_identifier,source_database_oid,target_system_identifier,
    target_database_oid,publication_name,slot_name,state
) SELECT $3,$4,$5,$6,$7,'tenant',$1,'iam_system_resources','iam',1,1,$2,$8,
    clock_timestamp(),$9,$10,$11,control.system_identifier,database_row.oid,
    control.system_identifier,database_row.oid,$12,$13,'copying'
FROM pg_catalog.pg_control_system() AS control
JOIN pg_catalog.pg_database AS database_row ON database_row.datname=current_database();
INSERT INTO platform.cutover_operation_relations (
    operation_id,run_id,ordinal,source_schema_oid,source_relation_oid,
    source_schema_name,source_relation_name,target_schema_oid,target_relation_oid,
    target_schema_name,target_relation_name,replica_identity
) VALUES ($3,$4,1,$14,$15,$16,$17,$18,$19,$20,$21,'primary_key');
INSERT INTO platform.cutover_operation_checkpoint_keys (
    operation_id,run_id,relation_ordinal,ordinal,checkpoint_key
) VALUES ($3,$4,1,1,'main');
INSERT INTO platform.cutover_operation_columns (
    operation_id,run_id,relation_ordinal,ordinal,source_attnum,target_attnum,
    transform_id,primary_key
) VALUES ($3,$4,1,1,1,1,'identity.uuid',true);
INSERT INTO platform.cutover_operation_procedures (
    operation_id,run_id,ordinal,relation_ordinal,procedure_kind,procedure_oid,
    procedure_schema,procedure_name,identity_arguments,procedure_signature,
    procedure_contract_id
) VALUES ($3,$4,1,1,'apply',$22,'platform','cutover_probe_apply_v1',
    'uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text',
    $23,'platform.cutover_probe.v1');
INSERT INTO platform.cutover_initial_copy_receipts (
    receipt_id,operation_id,run_id,manifest_sha256,snapshot_name,snapshot_lsn,
    relation_count,partition_count,row_count,root_sha256,completed_at
) VALUES (pg_catalog.gen_random_uuid(),$3,$4,$9,'snapshot-main','0/10',1,1,1,$24,
    pg_catalog.clock_timestamp());
INSERT INTO platform.cutover_change_journal (
    operation_id,run_id,scope_kind,scope_id,domain,cutover_generation,
    source_xid,source_session_user
) VALUES ($3,$4,'tenant',$1,'iam',1,'41'::xid8,'harpy_v1_runtime');`, pgx.QueryExecModeSimpleProtocol,
		artifacts.identity.scopeID, integrationOwnerID,
		artifacts.identity.operationID, artifacts.identity.runID, artifacts.identity.leaseID,
		integrationReleaseID, integrationEnvironment, artifacts.lease.ExpiresAt,
		artifacts.leaseArtifact.digest, artifacts.lease.PlanSHA256,
		artifacts.lease.TransformManifestSHA256, artifacts.lease.PublicationName,
		artifacts.lease.SlotName, relation.Source.SchemaOID, relation.Source.RelationOID,
		relation.Source.SchemaName, relation.Source.RelationName, relation.Target.SchemaOID,
		relation.Target.RelationOID, relation.Target.SchemaName, relation.Target.RelationName,
		relation.ApplyProcedureOID, relation.ApplyProcedureSignature,
		digestMarker('d'),
	)
	if err != nil {
		t.Fatalf("seed built cutover operation: %v", err)
	}
}

func leaseBuiltCutoverRoles(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase, expiresAt time.Time,
) {
	t.Helper()
	stamp := expiresAt.Format(time.RFC3339Nano)
	statement := fmt.Sprintf(`ALTER ROLE harpy_cutover_decoder LOGIN NOINHERIT
NOSUPERUSER NOCREATEDB NOCREATEROLE REPLICATION NOBYPASSRLS CONNECTION LIMIT 2
PASSWORD '%s' VALID UNTIL '%s';
ALTER ROLE harpy_adopter LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE
NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1 PASSWORD '%s' VALID UNTIL '%s';
GRANT SELECT ON legacy_source.allowed_rows TO harpy_cutover_decoder;`,
		builtCutoverPassword, stamp, builtCutoverPassword, stamp)
	if _, err := database.admin.Exec(ctx, statement); err != nil {
		t.Fatalf("lease built cutover roles: %v", err)
	}
}

func sealBuiltCutoverRoles(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, `ALTER ROLE harpy_cutover_decoder
NOLOGIN NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL;
ALTER ROLE harpy_adopter NOLOGIN NOREPLICATION CONNECTION LIMIT 0 PASSWORD NULL;`)
	if err != nil {
		t.Fatalf("seal built cutover roles: %v", err)
	}
	assertBuiltRoleConnectionDenied(t, ctx, database, "harpy_cutover_decoder", builtCutoverPassword)
	assertBuiltRoleConnectionDenied(t, ctx, database, "harpy_adopter", builtCutoverPassword)
}

func leaseBuiltFinalizerRole(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase, expiresAt time.Time,
) {
	t.Helper()
	statement := fmt.Sprintf(`ALTER ROLE harpy_cutover_finalizer LOGIN NOINHERIT
NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1
PASSWORD '%s' VALID UNTIL '%s'`, builtCutoverPassword, expiresAt.Format(time.RFC3339Nano))
	if _, err := database.admin.Exec(ctx, statement); err != nil {
		t.Fatalf("lease built finalizer role: %v", err)
	}
}

func builtCutoverRoleDSN(t *testing.T, dsn, role, password string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword(role, password)
	return parsed.String()
}
