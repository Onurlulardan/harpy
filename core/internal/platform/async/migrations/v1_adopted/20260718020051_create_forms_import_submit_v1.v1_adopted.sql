-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_import_submit_v1(
    p_tenant_id uuid,
    p_import_id uuid,
    p_job_id uuid,
    p_receipt_id uuid,
    p_principal_id uuid,
    p_authorization_version bigint,
    p_idempotency_key text,
    p_request_sha256 text,
    p_created_at timestamptz
)
RETURNS TABLE(
    resolved_import_id uuid,
    resolved_job_id uuid,
    resolved_receipt_id uuid,
    replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_authority_id uuid;
    v_group_id text;
    v_group_generation bigint;
    v_billing_generation bigint;
    v_forms_generation bigint;
    v_payload_ref text;
    v_payload_sha256 text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_tenant_id IS NULL OR p_principal_id IS NULL OR
       pg_catalog.uuid_extract_version(p_import_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_authorization_version NOT BETWEEN 1 AND 9007199254740991 OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_request_sha256 !~ '^[0-9a-f]{64}$' OR
       NOT pg_catalog.isfinite(p_created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Forms import submit input is invalid';
    END IF;
    IF NOT pg_catalog.pg_try_advisory_xact_lock(pg_catalog.hashtextextended(
        p_tenant_id::text || ':' || p_principal_id::text ||
        ':forms.records.import:' || p_idempotency_key, 0
    )) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6040',
            MESSAGE = 'Forms import idempotency is in progress';
    END IF;
    SELECT replay.import_id,replay.job_id,replay.receipt_id,replay.replayed
    INTO resolved_import_id,resolved_job_id,resolved_receipt_id,replayed
    FROM platform.platform_async_forms_import_replay_v1(
        p_tenant_id,p_principal_id,p_idempotency_key,p_request_sha256
    ) AS replay;
    IF FOUND THEN RETURN NEXT; RETURN; END IF;

    SELECT receipt.authority_id,receipt.authority_group_id,
           receipt.authority_group_generation,receipt.billing_generation,
           receipt.forms_generation
    INTO v_authority_id,v_group_id,v_group_generation,
         v_billing_generation,v_forms_generation
    FROM platform.billing_resolve_current_tenant_authority_receipt_v1(
        p_tenant_id
    ) AS receipt;
    IF NOT FOUND OR v_authority_id IS NULL OR v_group_id <> 'billing_forms' OR
       v_billing_generation < 1 OR v_forms_generation < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms import current authority is unavailable';
    END IF;

    v_payload_ref := 'forms-record-import/' || p_import_id::text;
    v_payload_sha256 := pg_catalog.encode(
        pg_catalog.sha256(pg_catalog.convert_to(v_payload_ref,'UTF8')),'hex'
    );
    INSERT INTO platform.jobs (
        id,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,
        writer_origin,kind,payload_revision,payload_ref,payload_sha256,
        dedupe_key,effect_class,priority,run_at,state,attempts,max_attempts,
        lease_generation,created_at,updated_at
    ) VALUES (
        p_job_id,p_tenant_id,'tenant',p_tenant_id,v_group_id,
        v_group_generation,'forms',v_forms_generation,'v2',
        'forms.record-import',1,v_payload_ref,v_payload_sha256,
        'forms.record-import:' || p_import_id::text,'database_transactional',
        50,p_created_at,'queued',0,8,0,p_created_at,p_created_at
    );
    INSERT INTO platform.idempotency_receipts (
        receipt_id,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,request_sha256,retention_class,
        principal_id,operation_id,idempotency_key,authorization_version,
        response_status,response_headers,http_result_ref,response_sha256,
        async_job_id,completed_at,retain_until
    ) VALUES (
        p_receipt_id,'http',p_tenant_id,'tenant',p_tenant_id,v_group_id,
        v_group_generation,'forms',v_forms_generation,'v2',p_request_sha256,
        'async_handle',p_principal_id,'forms.records.import',p_idempotency_key,
        p_authorization_version,202,'{}'::jsonb,v_payload_ref,v_payload_sha256,
        p_job_id,p_created_at,p_created_at + interval '91 days'
    );
    resolved_import_id := p_import_id;
    resolved_job_id := p_job_id;
    resolved_receipt_id := p_receipt_id;
    replayed := false;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_forms_import_submit_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,text,timestamptz
) FROM PUBLIC;
