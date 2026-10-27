-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_succeed_v1(
    p_tenant_id uuid,
    p_export_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_attempt_generation bigint,
    p_lease_owner_id uuid,
    p_effect_id uuid,
    p_effect_request_sha256 bytea,
    p_effect_receipt_id uuid,
    p_effect_result_sha256 bytea,
    p_provider_reference text,
    p_ciphertext_bytes bigint,
    p_completed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_receipt_id uuid;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_effect_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.uuid_extract_version(p_effect_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_effect_result_sha256) IS DISTINCT FROM 32 OR
       p_ciphertext_bytes NOT BETWEEN 1 AND 2147483648 OR
       NOT pg_catalog.isfinite(p_completed_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Forms export success input is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.kind='forms.record-export' AND job.payload_revision=1
      AND job.authority_domain='forms' AND job.writer_origin='v2'
      AND job.payload_ref='forms-record-export/' || p_export_id::text
      AND job.max_attempts=3 FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id FOR UPDATE;
    IF v_job.state IS DISTINCT FROM 'running' OR v_attempt.attempt_id IS NULL OR
       v_attempt.finished_at IS NOT NULL OR v_job.cancel_requested_at IS NOT NULL OR
       v_job.attempts IS DISTINCT FROM p_attempt_generation OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_until <= pg_catalog.clock_timestamp() OR
       v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.effect_request_sha256 IS DISTINCT FROM
           pg_catalog.encode(p_effect_request_sha256,'hex') OR
       p_completed_at <= v_job.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6047',
            MESSAGE = 'Forms export success lease is stale';
    END IF;
    v_receipt_id := platform.platform_async_forms_export_effect_receipt_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_effect_receipt_id,
        'succeeded',NULL,'OBJECT_UPLOAD_SUCCEEDED',p_provider_reference,
        p_effect_result_sha256,p_completed_at
    );
    UPDATE platform.job_attempts
    SET finished_at=p_completed_at,outcome='succeeded',
        effect_receipt_id=v_receipt_id
    WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET state='succeeded',lease_owner_id=NULL,
        lease_heartbeat_at=NULL,lease_until=NULL,last_error_class=NULL,
        last_error_code=NULL,last_error_sha256=NULL,
        retain_until=p_completed_at + interval '91 days',
        updated_at=p_completed_at,completed_at=p_completed_at
    WHERE id=p_job_id;
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
        used_ciphertext_bytes=used_ciphertext_bytes+p_ciphertext_bytes,
        revision=revision+1,updated_at=p_completed_at
    WHERE tenant_id=p_tenant_id
      AND reserved_ciphertext_bytes >= 2147483648
      AND used_ciphertext_bytes+p_ciphertext_bytes <= 21474836480;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6049',
            MESSAGE = 'Forms export success quota is inconsistent';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_succeed_v1(
    uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,bytea,text,bigint,timestamptz
) FROM PUBLIC;
