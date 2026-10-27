-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_platform_finish_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,
    p_failure_class text,p_failure_code text,p_failure_sha256 bytea,
    p_retryable boolean,p_provider_reference text,p_finished_at timestamptz
)
RETURNS TABLE(state text,effect_receipt_id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_attempt analytics_dashboard_export_platform_attempts%ROWTYPE;
    v_job_state text;
    v_receipt_id uuid;
    v_outcome text;
    v_domain_state text;
    v_quota_state text;
    v_exact boolean;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR p_lease_owner IS NULL OR
       (p_effect_id IS NULL) IS DISTINCT FROM
           (p_effect_request_sha256 IS NULL) OR
       (p_effect_id IS NOT NULL AND
        (pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
         pg_catalog.octet_length(p_effect_request_sha256) IS DISTINCT FROM 32)) OR
       p_failure_class IS NULL OR
       p_failure_class NOT IN ('transient','permanent','conflict','cancelled') OR
       p_failure_code IS NULL OR p_failure_code !~ '^[A-Z][A-Z0-9_]{1,63}$' OR
       pg_catalog.octet_length(p_failure_sha256) IS DISTINCT FROM 32 OR
       p_retryable IS NULL OR
       (p_retryable AND p_failure_class NOT IN ('transient','conflict')) OR
       (p_failure_class='cancelled' AND p_retryable) OR
       (p_provider_reference IS NOT NULL AND
        pg_catalog.octet_length(p_provider_reference) NOT BETWEEN 1 AND 512) OR
       p_finished_at IS NULL OR NOT pg_catalog.isfinite(p_finished_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export finish scope is invalid';
    END IF;
    SELECT finish.job_state,finish.effect_receipt_id
    INTO v_job_state,v_receipt_id
    FROM platform.platform_async_analytics_dashboard_export_finish_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_lease_owner,p_failure_class,p_failure_code,p_failure_sha256,
        p_retryable,p_provider_reference,p_finished_at
    ) AS finish;
    v_outcome:=CASE v_job_state WHEN 'queued' THEN 'requeued'
        WHEN 'dead' THEN 'dead' WHEN 'cancelled' THEN 'cancelled' END;
    v_domain_state:=CASE v_job_state WHEN 'queued' THEN 'queued'
        WHEN 'dead' THEN 'failed' WHEN 'cancelled' THEN 'cancelled' END;
    v_quota_state:=CASE WHEN v_job_state='queued' THEN 'none' ELSE 'released' END;
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=p_job_id FOR UPDATE;
    SELECT attempt.* INTO v_attempt
    FROM analytics_dashboard_export_platform_attempts AS attempt
    WHERE attempt.export_id=p_export_id
      AND attempt.attempt_generation=p_attempt_generation
      AND attempt.platform_attempt_id=p_attempt_id
      AND attempt.job_id=p_job_id AND attempt.lease_owner=p_lease_owner
    FOR UPDATE;
    IF v_export.export_id IS NULL OR v_attempt.export_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export finish attempt is stale';
    END IF;
    IF v_attempt.outcome IN ('requeued','dead','cancelled') THEN
        IF v_outcome IS NULL OR v_attempt.outcome IS DISTINCT FROM v_outcome OR
           v_attempt.platform_effect_receipt_id IS DISTINCT FROM v_receipt_id OR
           v_attempt.platform_effect_id IS DISTINCT FROM p_effect_id OR
           v_attempt.platform_effect_request_sha256 IS DISTINCT FROM
                p_effect_request_sha256 OR
           v_attempt.failure_class IS DISTINCT FROM p_failure_class OR
           v_attempt.failure_code IS DISTINCT FROM p_failure_code OR
           v_attempt.failure_sha256 IS DISTINCT FROM p_failure_sha256 OR
           v_attempt.finished_at IS DISTINCT FROM p_finished_at OR
           v_export.state IS DISTINCT FROM v_domain_state OR
           v_export.quota_state IS DISTINCT FROM v_quota_state OR
           v_export.platform_attempt_id IS NOT NULL OR
           v_export.platform_effect_id IS NOT NULL OR
           v_export.platform_effect_request_sha256 IS NOT NULL OR
           v_export.effect_started_at IS NOT NULL OR
           v_export.lease_owner IS NOT NULL OR v_export.lease_until IS NOT NULL OR
           v_export.updated_at IS DISTINCT FROM p_finished_at OR
           v_export.failure_code IS DISTINCT FROM
               (CASE WHEN v_job_state='dead' THEN p_failure_code END) THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export finish replay differs';
        END IF;
        IF v_attempt.platform_effect_id IS NOT NULL THEN
            SELECT receipt.outcome IS NOT DISTINCT FROM v_outcome AND
                   receipt.provider_request_id IS NOT DISTINCT FROM
                       p_provider_reference AND
                   receipt.completed_at IS NOT DISTINCT FROM p_finished_at AND
                   receipt.platform_effect_receipt_id IS NOT DISTINCT FROM
                       v_receipt_id AND
                   receipt.platform_effect_result_sha256 IS NULL INTO v_exact
            FROM analytics_dashboard_export_attempt_receipts AS receipt
            WHERE receipt.export_id=p_export_id
              AND receipt.attempt_generation=p_attempt_generation FOR UPDATE;
            IF NOT COALESCE(v_exact,false) THEN
                RAISE EXCEPTION USING ERRCODE='P6048',
                    MESSAGE='Analytics dashboard export finish receipt differs';
            END IF;
        END IF;
        state:=v_export.state;effect_receipt_id:=v_receipt_id;
        RETURN NEXT;RETURN;
    END IF;
    IF v_attempt.outcome IS DISTINCT FROM 'running' OR
       v_export.state IS DISTINCT FROM 'running' OR
       v_export.quota_state IS DISTINCT FROM 'held' OR
       v_export.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
       v_export.attempt_generation IS DISTINCT FROM p_attempt_generation OR
       v_export.lease_owner IS DISTINCT FROM p_lease_owner OR
       v_export.platform_effect_id IS DISTINCT FROM p_effect_id OR
       v_export.platform_effect_request_sha256 IS DISTINCT FROM
            p_effect_request_sha256 OR
       ((v_export.effect_started_at IS NULL) IS DISTINCT FROM
            (p_effect_id IS NULL)) OR
       p_finished_at<=v_export.updated_at OR
       (v_export.effect_started_at IS NOT NULL AND
        p_failure_class='cancelled') OR
       (p_failure_class='cancelled' AND
        v_export.cancellation_requested_at IS NULL) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export finish lease is stale';
    END IF;
    IF v_outcome IS NULL OR
       (v_export.effect_started_at IS NULL AND v_receipt_id IS NOT NULL) OR
       (v_export.effect_started_at IS NOT NULL AND v_receipt_id IS NULL) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export finish result is invalid';
    END IF;
    IF v_export.effect_started_at IS NOT NULL THEN
        UPDATE analytics_dashboard_export_attempt_receipts
        SET outcome=v_outcome,provider_request_id=p_provider_reference,
            completed_at=p_finished_at,
            platform_effect_receipt_id=v_receipt_id,
            platform_effect_result_sha256=NULL
        WHERE export_id=p_export_id
          AND attempt_generation=p_attempt_generation AND outcome='started';
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export failure receipt changed';
        END IF;
    END IF;
    UPDATE analytics_dashboard_export_platform_attempts
    SET outcome=v_outcome,platform_effect_receipt_id=v_receipt_id,
        platform_effect_result_sha256=NULL,failure_class=p_failure_class,
        failure_code=p_failure_code,failure_sha256=p_failure_sha256,
        finished_at=p_finished_at
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation AND outcome='running';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export finish attempt changed';
    END IF;
    UPDATE analytics_dashboard_exports
    SET state=v_domain_state,quota_state=v_quota_state,
        failure_code=CASE WHEN v_job_state='dead' THEN p_failure_code END,
        lease_owner=NULL,lease_until=NULL,platform_attempt_id=NULL,
        platform_effect_id=NULL,platform_effect_request_sha256=NULL,
        effect_started_at=NULL,object_key=NULL,object_version=NULL,key_id=NULL,
        wrapped_dek=NULL,wrap_nonce=NULL,header_sha256=NULL,
        ciphertext_sha256=NULL,plaintext_sha256=NULL,ciphertext_bytes=NULL,
        plaintext_bytes=NULL,row_count=NULL,filename=NULL,content_type=NULL,
        provider_request_id=NULL,row_revision=row_revision+1,
        updated_at=p_finished_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    state:=v_domain_state;effect_receipt_id:=v_receipt_id;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_finish_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,
    text,text,bytea,boolean,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_finish_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,
    text,text,bytea,boolean,text,timestamptz) TO forgecrud_runtime;
