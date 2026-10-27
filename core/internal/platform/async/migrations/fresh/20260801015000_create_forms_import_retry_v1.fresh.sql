-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_import_retry_v1(
    p_tenant_id uuid,p_import_id uuid,p_source_job_id uuid,
    p_retry_job_id uuid,p_receipt_id uuid,p_principal_id uuid,
    p_authorization_version bigint,p_idempotency_key text,
    p_request_sha256 text,p_created_at timestamptz)
RETURNS TABLE(resolved_job_id uuid,resolved_receipt_id uuid,
    retry_of_job_id uuid,replayed boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_existing_receipt platform.idempotency_receipts%ROWTYPE;
    v_source platform.jobs%ROWTYPE;
    v_retry platform.jobs%ROWTYPE;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_tenant_id IS NULL OR p_principal_id IS NULL OR
       pg_catalog.uuid_extract_version(p_import_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_source_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_retry_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_retry_job_id=p_source_job_id OR
       p_authorization_version NOT BETWEEN 1 AND 9007199254740991 OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 1 AND 128 OR
       p_request_sha256 !~ '^[0-9a-f]{64}$' OR
       NOT pg_catalog.isfinite(p_created_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Forms import retry input is invalid';
    END IF;
    IF NOT pg_catalog.pg_try_advisory_xact_lock(pg_catalog.hashtextextended(
        p_tenant_id::text || ':' || p_principal_id::text ||
        ':forms.records.import.resume:' || p_idempotency_key,0
    )) THEN
        RAISE EXCEPTION USING ERRCODE='P6040',
            MESSAGE='Forms import retry idempotency is in progress';
    END IF;
    SELECT receipt.* INTO v_existing_receipt
    FROM platform.idempotency_receipts AS receipt
    WHERE receipt.tenant_id=p_tenant_id
      AND receipt.principal_id=p_principal_id
      AND receipt.operation_id='forms.records.import.resume'
      AND receipt.idempotency_key=p_idempotency_key
    FOR SHARE;
    IF FOUND THEN
        IF v_existing_receipt.request_sha256<>p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='P6041',
                MESSAGE='Forms import retry idempotency key was reused';
        END IF;
        SELECT job.* INTO v_retry
        FROM platform.jobs AS job
        WHERE job.id=v_existing_receipt.async_job_id
          AND job.tenant_id=p_tenant_id
          AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
          AND job.authority_domain='forms' AND job.writer_origin='v2'
          AND job.kind='forms.record-import' AND job.payload_revision=1
          AND job.effect_class='database_transactional'
          AND job.effect_idempotency_key IS NULL AND job.max_attempts=8
          AND job.retry_of_job_id IS NOT NULL
          AND job.dedupe_key='forms.record-import.retry:' ||
              job.retry_of_job_id::text
        FOR SHARE;
        IF NOT FOUND OR v_retry.retry_of_job_id<>p_source_job_id OR
           v_existing_receipt.record_kind<>'http' OR
           v_existing_receipt.retention_class<>'async_handle' OR
           v_existing_receipt.scope_kind<>'tenant' OR
           v_existing_receipt.scope_id<>p_tenant_id OR
           v_existing_receipt.authority_group_id<>v_retry.authority_group_id OR
           v_existing_receipt.authority_group_generation<>
              v_retry.authority_group_generation OR
           v_existing_receipt.authority_domain<>v_retry.authority_domain OR
           v_existing_receipt.cutover_generation<>v_retry.cutover_generation OR
           v_existing_receipt.writer_origin<>v_retry.writer_origin OR
           v_existing_receipt.response_status<>202 OR
           v_existing_receipt.http_result_ref<>
              'forms-record-import/' || p_import_id::text OR
           v_existing_receipt.response_sha256<>v_retry.payload_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Forms import retry replay is inconsistent';
        END IF;
        PERFORM platform.platform_async_assert_current_authority_v1(
            v_retry.scope_kind,v_retry.scope_id,v_retry.authority_group_id,
            v_retry.authority_group_generation,v_retry.authority_domain,
            v_retry.cutover_generation);
        resolved_job_id:=v_retry.id; resolved_receipt_id:=v_existing_receipt.receipt_id;
        retry_of_job_id:=v_retry.retry_of_job_id; replayed:=true;
        RETURN NEXT;
        RETURN;
    END IF;
    SELECT job.* INTO v_source
    FROM platform.jobs AS job
    WHERE job.id=p_source_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_domain='forms' AND job.writer_origin='v2'
      AND job.kind='forms.record-import' AND job.payload_revision=1
      AND job.effect_class='database_transactional'
      AND job.effect_idempotency_key IS NULL
      AND ((job.retry_of_job_id IS NULL AND job.dedupe_key=
            'forms.record-import:' || p_import_id::text) OR
           (job.retry_of_job_id IS NOT NULL AND job.dedupe_key=
            'forms.record-import.retry:' || job.retry_of_job_id::text))
      AND job.payload_sha256=pg_catalog.encode(pg_catalog.sha256(
          pg_catalog.convert_to('forms-record-import/' || p_import_id::text,'UTF8')
      ),'hex')
      AND job.max_attempts=8
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='Forms import retry source was not found';
    END IF;
    IF v_source.state IS DISTINCT FROM 'dead' OR
       v_source.cancel_requested_at IS NOT NULL OR
       v_source.effect_started_at IS NOT NULL OR
       v_source.payload_ref IS DISTINCT FROM
          'forms-record-import/' || p_import_id::text OR
       v_source.payload_redacted_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6045',
            MESSAGE='Forms import job is not retryable';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.idempotency_receipts AS handle
        WHERE handle.tenant_id=p_tenant_id
          AND handle.scope_kind='tenant' AND handle.scope_id=p_tenant_id
          AND handle.principal_id=p_principal_id
          AND handle.operation_id IN (
              'forms.records.import','forms.records.import.resume'
          )
          AND handle.retention_class='async_handle'
          AND handle.async_job_id=v_source.id
          AND handle.http_result_ref=v_source.payload_ref
          AND handle.response_sha256=v_source.payload_sha256
          AND handle.authority_group_id=v_source.authority_group_id
          AND handle.authority_group_generation=
              v_source.authority_group_generation
          AND handle.authority_domain=v_source.authority_domain
          AND handle.cutover_generation=v_source.cutover_generation
          AND handle.writer_origin=v_source.writer_origin
          AND handle.retain_until>p_created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Forms import retry actor authority is invalid';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        v_source.scope_kind,v_source.scope_id,v_source.authority_group_id,
        v_source.authority_group_generation,v_source.authority_domain,
        v_source.cutover_generation
    );
    INSERT INTO platform.jobs(
        id,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,
        writer_origin,kind,payload_revision,payload_ref,payload_sha256,
        dedupe_key,effect_class,effect_idempotency_key,priority,run_at,
        state,attempts,max_attempts,lease_generation,retry_of_job_id,
        created_at,updated_at
    ) VALUES(
        p_retry_job_id,v_source.tenant_id,v_source.scope_kind,v_source.scope_id,
        v_source.authority_group_id,v_source.authority_group_generation,
        v_source.authority_domain,v_source.cutover_generation,'v2',
        v_source.kind,v_source.payload_revision,v_source.payload_ref,
        v_source.payload_sha256,'forms.record-import.retry:' || v_source.id::text,
        v_source.effect_class,v_source.effect_idempotency_key,v_source.priority,
        p_created_at,'queued',0,v_source.max_attempts,0,v_source.id,
        p_created_at,p_created_at
    ) ON CONFLICT ON CONSTRAINT jobs_dedupe_unique DO NOTHING
    RETURNING * INTO v_retry;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6045',
            MESSAGE='Forms import retry conflicted';
    END IF;
    replayed:=false;
    INSERT INTO platform.idempotency_receipts(
        receipt_id,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,request_sha256,retention_class,
        principal_id,operation_id,idempotency_key,authorization_version,
        response_status,response_headers,http_result_ref,response_sha256,
        async_job_id,completed_at,retain_until
    ) VALUES(
        p_receipt_id,'http',p_tenant_id,'tenant',p_tenant_id,
        v_retry.authority_group_id,v_retry.authority_group_generation,
        v_retry.authority_domain,v_retry.cutover_generation,'v2',
        p_request_sha256,'async_handle',p_principal_id,
        'forms.records.import.resume',p_idempotency_key,
        p_authorization_version,202,'{}'::jsonb,
        'forms-record-import/' || p_import_id::text,v_retry.payload_sha256,
        v_retry.id,p_created_at,p_created_at+interval '91 days'
    );
    resolved_job_id:=v_retry.id;
    resolved_receipt_id:=p_receipt_id;
    retry_of_job_id:=v_retry.retry_of_job_id;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_forms_import_retry_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,text,timestamptz
) FROM PUBLIC;
