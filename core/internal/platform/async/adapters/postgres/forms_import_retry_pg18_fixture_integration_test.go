//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

const formsImportRetryPGFixture = `
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime LOGIN PASSWORD 'retry_runtime_only';
GRANT harpy_runtime TO harpy_v2_runtime;
CREATE TABLE platform.jobs(
 id uuid PRIMARY KEY,tenant_id uuid,scope_kind text,scope_id uuid,
 authority_group_id text,authority_group_generation bigint,
 authority_domain text,cutover_generation bigint,writer_origin text,
 kind text,payload_revision integer,payload_ref text,payload_sha256 text,
 payload_redacted_at timestamptz,dedupe_key text,effect_class text,
 effect_idempotency_key text,priority integer,run_at timestamptz,state text,
 attempts integer,max_attempts integer,lease_generation bigint,
 retry_of_job_id uuid,cancel_requested_at timestamptz,
 effect_started_at timestamptz,lease_owner_id uuid,
 lease_heartbeat_at timestamptz,lease_until timestamptz,
 last_error_class text,last_error_code text,last_error_sha256 text,
 retain_until timestamptz,completed_at timestamptz,
 created_at timestamptz,updated_at timestamptz,
 CONSTRAINT jobs_dedupe_unique UNIQUE(
  scope_kind,scope_id,authority_group_id,authority_domain,kind,dedupe_key
 )
);
CREATE TABLE platform.job_attempts(
 attempt_id uuid PRIMARY KEY,job_id uuid,scope_kind text,scope_id uuid,
 attempt integer,lease_generation bigint,lease_owner_id uuid,
 started_at timestamptz,finished_at timestamptz,outcome text
);
CREATE TABLE platform.idempotency_receipts(
 receipt_id uuid PRIMARY KEY,record_kind text,tenant_id uuid,scope_kind text,
 scope_id uuid,authority_group_id text,authority_group_generation bigint,
 authority_domain text,cutover_generation bigint,writer_origin text,
 request_sha256 text,retention_class text,principal_id uuid,operation_id text,
 idempotency_key text,authorization_version bigint,response_status integer,
 response_headers jsonb,http_result_ref text,response_sha256 text,
 async_job_id uuid,completed_at timestamptz,retain_until timestamptz,
 UNIQUE(tenant_id,principal_id,operation_id,idempotency_key)
);
CREATE FUNCTION platform.platform_async_assert_current_authority_v1(
 p_scope_kind text,p_scope_id uuid,p_group_id text,p_group_generation bigint,
 p_domain text,p_generation bigint
) RETURNS boolean LANGUAGE plpgsql AS $function$
BEGIN
 IF p_scope_kind<>'tenant' OR p_scope_id IS NULL OR
    p_group_id<>'billing_forms' OR p_group_generation<>3 OR
    p_domain<>'forms' OR p_generation<>4 THEN
  RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='authority is stale';
 END IF;
 RETURN true;
END;
$function$;
`

const formsImportRetryWrapper = `
CREATE FUNCTION public.retry_forms_import(
 p_tenant_id uuid,p_import_id uuid,p_source_job_id uuid,
 p_retry_job_id uuid,p_receipt_id uuid,p_principal_id uuid,
 p_authorization_version bigint,p_idempotency_key text,
 p_request_sha256 text,p_created_at timestamptz
) RETURNS TABLE(
 resolved_job_id uuid,resolved_receipt_id uuid,
 retry_of_job_id uuid,replayed boolean
) LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $function$
 SELECT * FROM platform.platform_async_forms_import_retry_v1(
  p_tenant_id,p_import_id,p_source_job_id,p_retry_job_id,p_receipt_id,
  p_principal_id,p_authorization_version,p_idempotency_key,
  p_request_sha256,p_created_at
 );
$function$;
REVOKE ALL ON FUNCTION public.retry_forms_import(
 uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.retry_forms_import(
 uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,text,timestamptz
) TO harpy_v2_runtime;
`

func prepareFormsImportRetryPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, formsImportRetryPGFixture); err != nil {
		t.Fatalf("prepare Forms import retry fixture: %v", err)
	}
	if _, err := database.ExecContext(ctx, formsImportRetryPriorFunctions); err != nil {
		t.Fatalf("prepare prior Forms import routines: %v", err)
	}
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	suffixes := []string{
		"20260801015000_create_forms_import_retry_v1.fresh.sql",
		"20260801015001_allow_forms_import_retry_read_v1.fresh.sql",
		"20260801015002_allow_forms_import_retry_control_v1.fresh.sql",
		"20260801015003_allow_forms_import_retry_current_lease_v1.fresh.sql",
		"20260801015004_allow_forms_import_retry_complete_v1.fresh.sql",
	}
	for _, suffix := range suffixes {
		found := false
		for _, document := range documents {
			if !strings.HasSuffix(document.Path(), suffix) {
				continue
			}
			found = true
			if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
				t.Fatalf("apply %s: %v", suffix, err)
			}
			break
		}
		if !found {
			t.Fatalf("Forms import retry migration %s is missing", suffix)
		}
	}
	if _, err := database.ExecContext(ctx,
		formsImportRetryWrapper+formsImportRetryLineageWrappers); err != nil {
		t.Fatalf("create Forms import retry wrappers: %v", err)
	}
}

func seedFormsImportRetrySource(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenantID, importID, jobID, actorID, receiptID uuid.UUID,
	state string, cancelled bool,
) {
	t.Helper()
	observedAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	var cancellation any
	if cancelled {
		cancellation = observedAt
	}
	_, err := database.ExecContext(ctx, `
INSERT INTO platform.jobs(
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
 priority,run_at,state,attempts,max_attempts,lease_generation,
 cancel_requested_at,created_at,updated_at
) VALUES($1,$2,'tenant',$2,'billing_forms',3,'forms',4,'v2',
 'forms.record-import',1,'forms-record-import/'||$3::text,
 encode(sha256(convert_to('forms-record-import/'||$3::text,'UTF8')),'hex'),

 'forms.record-import:'||$3::text,'database_transactional',50,$4,$5,8,8,8,
 $6,$4,$4)`, jobID, tenantID, importID, observedAt, state, cancellation)
	if err != nil {
		t.Fatalf("seed Forms import retry job: %v", err)
	}
	_, err = database.ExecContext(ctx, `
INSERT INTO platform.idempotency_receipts(
 receipt_id,record_kind,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 request_sha256,retention_class,principal_id,operation_id,idempotency_key,
 authorization_version,response_status,response_headers,http_result_ref,
 response_sha256,async_job_id,completed_at,retain_until
) SELECT $1,'http',$2,'tenant',$2,'billing_forms',3,'forms',4,'v2',
 repeat('a',64),'async_handle',$3,'forms.records.import',
 'submit:'||job.id::text,1,202,
 '{}'::jsonb,job.payload_ref,job.payload_sha256,job.id,
 $4::timestamptz,$4::timestamptz+interval '91 days'
 FROM platform.jobs AS job WHERE job.id=$5`,
		receiptID, tenantID, actorID, observedAt, jobID)
	if err != nil {
		t.Fatalf("seed Forms import retry source: %v", err)
	}
}
