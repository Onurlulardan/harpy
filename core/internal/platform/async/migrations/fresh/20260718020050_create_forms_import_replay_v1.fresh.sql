-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_import_replay_v1(
    p_tenant_id uuid,
    p_principal_id uuid,
    p_idempotency_key text,
    p_request_sha256 text
)
RETURNS TABLE(
    import_id uuid,
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
    v_import_text text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms import replay authority is invalid';
    END IF;
    SELECT candidate.* INTO v_receipt
    FROM platform.idempotency_receipts AS candidate
    WHERE candidate.record_kind = 'http'
      AND candidate.tenant_id = p_tenant_id
      AND candidate.principal_id = p_principal_id
      AND candidate.operation_id = 'forms.records.import'
      AND candidate.idempotency_key = p_idempotency_key
    FOR SHARE;
    IF NOT FOUND THEN RETURN; END IF;
    IF v_receipt.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6041',
            MESSAGE = 'Forms import idempotency key was reused';
    END IF;
    SELECT candidate.* INTO v_job
    FROM platform.jobs AS candidate
    WHERE candidate.id = v_receipt.async_job_id
      AND candidate.tenant_id = p_tenant_id
      AND candidate.scope_kind = 'tenant'
      AND candidate.scope_id = p_tenant_id
      AND candidate.authority_domain = 'forms'
      AND candidate.writer_origin = 'v2'
      AND candidate.kind = 'forms.record-import'
      AND candidate.payload_revision = 1
      AND candidate.effect_class = 'database_transactional'
      AND candidate.effect_idempotency_key IS NULL
      AND candidate.max_attempts = 8
    FOR SHARE;
    IF NOT FOUND OR v_receipt.retention_class IS DISTINCT FROM 'async_handle' OR
       v_receipt.response_status IS DISTINCT FROM 202 OR
       v_receipt.http_result_ref NOT LIKE 'forms-record-import/%' OR
       v_receipt.response_sha256 IS DISTINCT FROM pg_catalog.encode(
           pg_catalog.sha256(pg_catalog.convert_to(v_receipt.http_result_ref, 'UTF8')),
           'hex'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6047',
            MESSAGE = 'Forms import replay evidence is inconsistent';
    END IF;
    v_import_text := pg_catalog.substring(
        v_receipt.http_result_ref,
        pg_catalog.length('forms-record-import/') + 1
    );
    BEGIN
        import_id := v_import_text::uuid;
    EXCEPTION WHEN invalid_text_representation THEN
        RAISE EXCEPTION USING ERRCODE = 'P6047',
            MESSAGE = 'Forms import replay handle is invalid';
    END;
    IF v_job.dedupe_key IS DISTINCT FROM
           'forms.record-import:' || import_id::text OR
       v_job.payload_ref IS DISTINCT FROM
           'forms-record-import/' || import_id::text THEN
        RAISE EXCEPTION USING ERRCODE = 'P6047',
            MESSAGE = 'Forms import replay job differs';
    END IF;
    job_id := v_job.id;
    receipt_id := v_receipt.receipt_id;
    replayed := true;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_forms_import_replay_v1(
    uuid,uuid,text,text
) FROM PUBLIC;
