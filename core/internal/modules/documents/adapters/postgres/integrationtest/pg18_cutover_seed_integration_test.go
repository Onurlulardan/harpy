//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

const (
	documentsCutoverTenantID   = "018f0000-0000-7000-8000-000000000001"
	documentsCutoverOperation  = "018f0000-0000-7000-8000-000000000010"
	documentsCutoverRun        = "018f0000-0000-7000-8000-000000000011"
	documentsAdoptionRun       = "018f0000-0000-7000-8000-000000000020"
	documentsSourceDocument    = "018f0000-0000-7000-8000-000000000021"
	documentsSourceVersion     = "018f0000-0000-7000-8000-000000000022"
	documentsTargetFolder      = "018f0000-0000-7000-8000-000000000023"
	documentsTargetItem        = "018f0000-0000-7000-8000-000000000024"
	documentsTargetVersion     = "018f0000-0000-7000-8000-000000000025"
	documentsCopyJob           = "018f0000-0000-7000-8000-000000000026"
	documentsCutoverManifest   = "1111111111111111111111111111111111111111111111111111111111111111"
	documentsSourceContentHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	documentsSourceMetaHash    = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func seedDocumentsCutoverOperation(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	requestedBy := "018f0000-0000-7000-8000-000000000002"
	fixtures := []struct {
		statement string
		arguments []any
	}{
		{`INSERT INTO platform.tenants(id) VALUES ($1)`,
			[]any{documentsCutoverTenantID}},
		{`INSERT INTO platform.domain_cutovers (
 scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference)
VALUES ('tenant',$1,'documents','shadow','v1',13,$2,'Documents PG18','documents-pg18')`,
			[]any{documentsCutoverTenantID, requestedBy}},
		{`INSERT INTO platform.cutover_groups (
 scope_kind,scope_id,group_id,state,writer,generation,requested_by,reason,change_reference)
VALUES ('tenant',$1,'documents','shadow','v1',13,$2,'Documents PG18','documents-pg18')`,
			[]any{documentsCutoverTenantID, requestedBy}},
		{`INSERT INTO platform.cutover_group_members
 (scope_kind,scope_id,group_id,ordinal,domain)
VALUES ('tenant',$1,'documents',1,'documents')`,
			[]any{documentsCutoverTenantID}},
		{`INSERT INTO platform.cutover_operations (
 operation_id,run_id,lease_id,release_id,environment_id,scope_kind,scope_id,
 group_id,domain,cutover_generation,lease_generation,lease_owner,
 lease_expires_at,last_heartbeat_at,lease_manifest_sha256,plan_sha256,
 transform_manifest_sha256,source_system_identifier,source_database_oid,
 target_system_identifier,target_database_oid,publication_name,slot_name,state,updated_at)
VALUES ($3,$4,$5,$6,$7,'tenant',$1,'documents','documents',13,1,$2,
 '2099-01-01 00:00:00+00',pg_catalog.clock_timestamp(),$8,$9,$10,
 1,1,1,1,'documents_publication','documents_slot','copying',
 pg_catalog.clock_timestamp())`, []any{
			documentsCutoverTenantID, requestedBy, documentsCutoverOperation,
			documentsCutoverRun, "018f0000-0000-7000-8000-000000000012",
			"018f0000-0000-7000-8000-000000000013",
			"018f0000-0000-7000-8000-000000000014", documentsCutoverManifest,
			strings.Repeat("2", 64), strings.Repeat("3", 64),
		}},
		{`INSERT INTO platform.cutover_operation_relations (
 operation_id,run_id,ordinal,source_schema_oid,source_relation_oid,
 source_schema_name,source_relation_name,target_schema_oid,target_relation_oid,
 target_schema_name,target_relation_name,replica_identity)
VALUES ($1,$2,1,100,101,'legacy_documents','documents',102,103,$3,
 'documents_v1_adoption_stage','primary_key')`,
			[]any{documentsCutoverOperation, documentsCutoverRun, schema}},
		{`INSERT INTO platform.cutover_operation_checkpoint_keys
 (operation_id,run_id,relation_ordinal,ordinal,checkpoint_key)
VALUES ($1,$2,1,1,'folder'),($1,$2,1,2,'item'),
       ($1,$2,1,3,'version'),($1,$2,1,4,'inventory')`,
			[]any{documentsCutoverOperation, documentsCutoverRun}},
		{`INSERT INTO platform.cutover_operation_procedures (
 operation_id,run_id,ordinal,relation_ordinal,procedure_kind,procedure_oid,
 procedure_schema,procedure_name,identity_arguments,procedure_signature,
 procedure_contract_id)
SELECT $1,$2,1,1,'apply',procedure_row.oid,$3::text,
 'documents_v1_adoption_apply_v1',
 'uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text',
 pg_catalog.format('%I.documents_v1_adoption_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)',$3::text),
 'documents.v1_adoption.v1'
FROM pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=procedure_row.pronamespace
WHERE namespace.nspname::text=$3::text
  AND procedure_row.proname='documents_v1_adoption_apply_v1'`,
			[]any{documentsCutoverOperation, documentsCutoverRun, schema}},
	}
	for _, fixture := range fixtures {
		if _, err := database.ExecContext(ctx, fixture.statement,
			fixture.arguments...); err != nil {
			t.Fatal(err)
		}
	}
}
