-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_export_reconcile_effect_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_lease_owner_id uuid,p_lease_generation bigint,p_resource_kind text,
    p_effect_id uuid,p_receipt_id uuid,p_request_sha256 bytea,
    p_started_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_cancelled boolean;
    v_request text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_effect_id=p_receipt_id OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       NOT pg_catalog.isfinite(p_started_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export reconciliation effect is invalid';
    END IF;
    SELECT platform.platform_async_iam_export_reconcile_lease_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM export reconciliation cancellation was requested';
    END IF;
    IF EXISTS(
        SELECT 1 FROM platform.idempotency_receipts AS receipt
        WHERE receipt.receipt_id=p_receipt_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation receipt already exists';
    END IF;
    v_request := pg_catalog.encode(p_request_sha256,'hex');
    SELECT job.* INTO STRICT v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id FOR UPDATE;
    SELECT attempt.* INTO STRICT v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id FOR UPDATE;
    IF v_attempt.effect_started_at IS NOT NULL THEN
        IF v_attempt.effect_id=p_effect_id AND
           v_attempt.effect_request_sha256=v_request AND
           v_attempt.effect_started_at=p_started_at AND
           v_job.effect_started_at=p_started_at THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation effect replay differs';
    END IF;
    IF p_effect_id<>p_attempt_id OR p_started_at<v_attempt.started_at OR
       p_started_at<=v_job.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export reconciliation effect lease is stale';
    END IF;
    UPDATE platform.job_attempts SET effect_started_at=p_started_at,
        effect_id=p_effect_id,effect_request_sha256=v_request
    WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET effect_started_at=p_started_at,
        updated_at=p_started_at WHERE id=p_job_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_export_reconcile_effect_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz
) FROM PUBLIC;
