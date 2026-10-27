-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_expired_finalize_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,p_observed_at timestamptz
)
RETURNS TABLE(final_state text,finalized_at timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_attempt analytics_dashboard_export_platform_attempts%ROWTYPE;
    v_final_state text;
    v_finalized_at timestamptz;
    v_failure_sha256 bytea;
    v_outcome text;
    v_domain_state text;
    v_quota_state text;
    v_failure_class text;
    v_failure_code text;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id='00000000-0000-0000-0000-000000000000'::uuid OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export expiry scope is invalid';
    END IF;
    SELECT result.final_state,result.finalized_at,result.failure_sha256
    INTO STRICT v_final_state,v_finalized_at,v_failure_sha256
    FROM platform.platform_async_analytics_dashboard_export_expired_finalize_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_lease_owner,p_effect_id,p_effect_request_sha256,p_observed_at
    ) AS result;
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
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export expiry evidence is unavailable';
    END IF;
    IF v_final_state='outcome_unknown' THEN
        IF p_effect_id IS NULL OR v_failure_sha256 IS NOT NULL OR
           v_attempt.platform_effect_id IS DISTINCT FROM p_effect_id OR
           v_attempt.platform_effect_request_sha256 IS DISTINCT FROM
                p_effect_request_sha256 OR
           v_attempt.effect_started_at IS NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export unknown expiry result is invalid';
        END IF;
        IF v_attempt.outcome='unknown' THEN
            IF v_attempt.finished_at IS DISTINCT FROM v_finalized_at OR
               v_export.state IS DISTINCT FROM 'running' OR
               v_export.quota_state IS DISTINCT FROM 'held' OR
               v_export.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
               v_export.platform_effect_id IS DISTINCT FROM p_effect_id OR
               v_export.platform_effect_request_sha256 IS DISTINCT FROM p_effect_request_sha256 OR
               v_export.effect_started_at IS DISTINCT FROM
                    v_attempt.effect_started_at OR
               v_export.lease_owner IS NOT NULL OR
               v_export.failure_code IS DISTINCT FROM 'ARTIFACT_OUTCOME_UNKNOWN' THEN
                RAISE EXCEPTION USING ERRCODE='P6048',
                    MESSAGE='Analytics dashboard export unknown expiry replay differs';
            END IF;
            final_state:=v_final_state;finalized_at:=v_finalized_at;
            RETURN NEXT;RETURN;
        END IF;
        IF v_attempt.outcome IS DISTINCT FROM 'running' OR
           v_export.state IS DISTINCT FROM 'running' OR
           v_export.quota_state IS DISTINCT FROM 'held' OR
           v_export.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
           v_export.attempt_generation IS DISTINCT FROM p_attempt_generation OR
           v_export.lease_owner IS DISTINCT FROM p_lease_owner OR
           v_export.platform_effect_id IS DISTINCT FROM p_effect_id OR
           v_export.platform_effect_request_sha256 IS DISTINCT FROM p_effect_request_sha256 OR
           v_export.effect_started_at IS DISTINCT FROM v_attempt.effect_started_at OR
           v_finalized_at<v_export.updated_at THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export unknown expiry evidence differs';
        END IF;
        UPDATE analytics_dashboard_export_attempt_receipts
        SET outcome='unknown',completed_at=v_finalized_at
        WHERE export_id=p_export_id
          AND attempt_generation=p_attempt_generation AND outcome='started';
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export unknown expiry receipt differs';
        END IF;
        UPDATE analytics_dashboard_export_platform_attempts
        SET outcome='unknown',failure_code='JOB_OUTCOME_UNKNOWN',
            finished_at=v_finalized_at
        WHERE export_id=p_export_id AND attempt_generation=p_attempt_generation;
        UPDATE analytics_dashboard_exports
        SET failure_code='ARTIFACT_OUTCOME_UNKNOWN',lease_owner=NULL,
            lease_until=NULL,row_revision=row_revision+1,
            updated_at=v_finalized_at
        WHERE analytics_dashboard_exports.export_id=p_export_id;
        final_state:=v_final_state;finalized_at:=v_finalized_at;
        RETURN NEXT;RETURN;
    END IF;
    v_outcome:=CASE v_final_state WHEN 'queued' THEN 'requeued'
        WHEN 'dead' THEN 'dead' WHEN 'cancelled' THEN 'cancelled' END;
    v_domain_state:=CASE v_final_state WHEN 'queued' THEN 'queued'
        WHEN 'dead' THEN 'failed' WHEN 'cancelled' THEN 'cancelled' END;
    v_quota_state:=CASE WHEN v_final_state='queued' THEN 'none'
        ELSE 'released' END;
    v_failure_class:=CASE WHEN v_final_state='cancelled' THEN 'cancelled'
        ELSE 'transient' END;
    v_failure_code:=CASE WHEN v_final_state='cancelled' THEN 'EXPORT_CANCELLED'
        ELSE 'JOB_LEASE_EXPIRED' END;
    IF v_outcome IS NULL OR p_effect_id IS NOT NULL OR
       pg_catalog.octet_length(v_failure_sha256) IS DISTINCT FROM 32 OR
       v_failure_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export pre-effect expiry result is invalid';
    END IF;
    IF v_attempt.outcome IS DISTINCT FROM 'running' THEN
        IF v_attempt.outcome IS DISTINCT FROM v_outcome OR
           v_attempt.finished_at IS DISTINCT FROM v_finalized_at OR
           v_attempt.failure_class IS DISTINCT FROM v_failure_class OR
           v_attempt.failure_code IS DISTINCT FROM v_failure_code OR
           v_attempt.failure_sha256 IS DISTINCT FROM v_failure_sha256 OR
           v_attempt.platform_effect_id IS NOT NULL OR
           v_attempt.platform_effect_request_sha256 IS NOT NULL OR
           v_attempt.effect_started_at IS NOT NULL OR
           v_export.state IS DISTINCT FROM v_domain_state OR
           v_export.quota_state IS DISTINCT FROM v_quota_state OR
           v_export.platform_attempt_id IS NOT NULL OR
           v_export.platform_effect_id IS NOT NULL OR
           v_export.platform_effect_request_sha256 IS NOT NULL OR
           v_export.effect_started_at IS NOT NULL OR
           v_export.lease_owner IS NOT NULL OR v_export.lease_until IS NOT NULL OR
           v_export.updated_at IS DISTINCT FROM v_finalized_at THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export expiry replay differs';
        END IF;
        final_state:=v_final_state;finalized_at:=v_finalized_at;
        RETURN NEXT;RETURN;
    END IF;
    IF v_export.state IS DISTINCT FROM 'running' OR
       v_export.quota_state IS DISTINCT FROM 'held' OR
       v_export.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
       v_export.attempt_generation IS DISTINCT FROM p_attempt_generation OR
       v_export.lease_owner IS DISTINCT FROM p_lease_owner OR
       v_attempt.platform_effect_id IS NOT NULL OR
       v_attempt.platform_effect_request_sha256 IS NOT NULL OR
       v_attempt.effect_started_at IS NOT NULL OR
       v_export.effect_started_at IS NOT NULL OR
       v_finalized_at<v_export.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export pre-effect expiry evidence differs';
    END IF;
    UPDATE analytics_dashboard_export_platform_attempts
    SET outcome=v_outcome,failure_class=v_failure_class,
        failure_code=v_failure_code,failure_sha256=v_failure_sha256,
        finished_at=v_finalized_at
    WHERE export_id=p_export_id AND attempt_generation=p_attempt_generation;
    UPDATE analytics_dashboard_exports
    SET state=v_domain_state,quota_state=v_quota_state,
        failure_code=CASE WHEN v_final_state='dead' THEN v_failure_code END,
        lease_owner=NULL,lease_until=NULL,platform_attempt_id=NULL,
        platform_effect_id=NULL,platform_effect_request_sha256=NULL,
        row_revision=row_revision+1,updated_at=v_finalized_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    final_state:=v_final_state;finalized_at:=v_finalized_at;RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_expired_finalize_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_expired_finalize_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz
) TO forgecrud_runtime;
