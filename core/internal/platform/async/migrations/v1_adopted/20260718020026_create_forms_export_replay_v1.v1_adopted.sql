-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_replay_v1(
    p_tenant_id uuid,
    p_principal_id uuid,
    p_idempotency_key text,
    p_request_sha256 text
)
RETURNS TABLE(
    export_id uuid,
    job_id uuid,
    receipt_id uuid,
    replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_receipt platform.idempotency_receipts%ROWTYPE;
    v_job platform.jobs%ROWTYPE;
    v_export_text text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms export replay authority is invalid';
    END IF;

    SELECT candidate.* INTO v_receipt
    FROM platform.idempotency_receipts AS candidate
    WHERE candidate.record_kind = 'http'
      AND candidate.tenant_id = p_tenant_id
      AND candidate.principal_id = p_principal_id
      AND candidate.operation_id = 'forms.records.export'
      AND candidate.idempotency_key = p_idempotency_key
    FOR SHARE;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    IF v_receipt.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6031',
            MESSAGE = 'Forms export idempotency key was reused';
    END IF;

    SELECT candidate.* INTO v_job
    FROM platform.jobs AS candidate
    WHERE candidate.id = v_receipt.async_job_id
      AND candidate.tenant_id = p_tenant_id
      AND candidate.scope_kind = 'tenant'
      AND candidate.scope_id = p_tenant_id
      AND candidate.authority_domain = 'forms'
      AND candidate.writer_origin = 'v2'
      AND candidate.kind = 'forms.record-export'
      AND candidate.payload_revision = 1
      AND candidate.max_attempts = 3
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6037',
            MESSAGE = 'Forms export replay job is missing';
    END IF;
    IF v_receipt.retention_class IS DISTINCT FROM 'async_handle' OR
       v_receipt.response_status IS DISTINCT FROM 202 OR
       v_receipt.http_result_ref IS NULL OR
       v_receipt.http_result_ref NOT LIKE 'forms-record-export/%' OR
       v_receipt.response_sha256 IS DISTINCT FROM pg_catalog.encode(
           pg_catalog.sha256(pg_catalog.convert_to(v_receipt.http_result_ref, 'UTF8')),
           'hex'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6037',
            MESSAGE = 'Forms export replay evidence is inconsistent';
    END IF;
    v_export_text := pg_catalog.substring(
        v_receipt.http_result_ref,
        pg_catalog.length('forms-record-export/') + 1
    );
    BEGIN
        export_id := v_export_text::uuid;
    EXCEPTION WHEN invalid_text_representation THEN
        RAISE EXCEPTION USING ERRCODE = 'P6037',
            MESSAGE = 'Forms export replay handle is invalid';
    END;
    IF v_job.dedupe_key IS DISTINCT FROM
           'forms.record-export:' || export_id::text OR
       v_job.effect_idempotency_key IS DISTINCT FROM
           'forms.record-export:' || export_id::text OR
       NOT (v_job.payload_ref = 'forms-record-export/' || export_id::text OR
            (v_job.payload_ref IS NULL AND
             v_job.payload_redacted_at IS NOT NULL)) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6037',
            MESSAGE = 'Forms export replay job differs';
    END IF;
    job_id := v_job.id;
    receipt_id := v_receipt.receipt_id;
    replayed := true;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_replay_v1(
    uuid,uuid,text,text
) FROM PUBLIC;
