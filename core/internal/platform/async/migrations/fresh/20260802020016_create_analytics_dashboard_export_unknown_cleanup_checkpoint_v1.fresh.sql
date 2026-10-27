-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_unknown_cleanup_checkpoint_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_attempt_generation bigint,p_original_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_cleanup_generation bigint,
    p_cleanup_lease_owner_id uuid,p_observed_at timestamptz
)
RETURNS TABLE(lease_until timestamptz,cancellation_requested boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_claim platform.analytics_dashboard_export_unknown_cleanup_claims%ROWTYPE;
    v_request text;
    v_checkpoint_at timestamptz;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export cleanup checkpoint authority is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_original_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_request_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       p_cleanup_generation<1 OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       NOT pg_catalog.isfinite(p_observed_at) OR
       p_observed_at>pg_catalog.clock_timestamp()+interval '5 seconds' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export cleanup checkpoint input is invalid';
    END IF;
    v_request:=pg_catalog.encode(p_request_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_domain='analytics' AND job.writer_origin='v2'
      AND job.kind='analytics.dashboard.export.render'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=3 AND job.payload_ref=p_export_id::text
      AND job.state='dead' AND job.last_error_class IS NULL
      AND job.last_error_code='JOB_OUTCOME_UNKNOWN' FOR SHARE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_original_lease_owner_id
      AND attempt.outcome='dead' AND attempt.error_class IS NULL
      AND attempt.error_code='JOB_OUTCOME_UNKNOWN'
      AND attempt.effect_id=p_effect_id
      AND attempt.effect_request_sha256=v_request
      AND attempt.effect_started_at IS NOT NULL
      AND attempt.effect_receipt_id IS NULL
      AND attempt.finished_at=v_job.completed_at FOR SHARE;
    IF v_job.id IS NULL OR v_attempt.attempt_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export cleanup checkpoint evidence differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    SELECT claim.* INTO v_claim
    FROM platform.analytics_dashboard_export_unknown_cleanup_claims AS claim
    WHERE claim.job_id=p_job_id AND claim.tenant_id=p_tenant_id
      AND claim.export_id=p_export_id AND claim.attempt_id=p_attempt_id
      AND claim.attempt_generation=p_attempt_generation
      AND claim.original_lease_owner_id=p_original_lease_owner_id
      AND claim.effect_id=p_effect_id AND claim.request_sha256=v_request
    FOR UPDATE;
    IF v_claim.job_id IS NULL OR
       v_claim.claim_generation<>p_cleanup_generation OR
       v_claim.lease_owner_id<>p_cleanup_lease_owner_id OR
       v_claim.lease_until<=pg_catalog.clock_timestamp() OR EXISTS (
        SELECT 1 FROM platform.analytics_dashboard_export_unknown_resolutions resolution
        WHERE resolution.job_id=p_job_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export cleanup checkpoint lease is stale';
    END IF;
    v_checkpoint_at:=GREATEST(pg_catalog.clock_timestamp(),p_observed_at,
        v_claim.updated_at+interval '1 microsecond');
    UPDATE platform.analytics_dashboard_export_unknown_cleanup_claims
    SET lease_heartbeat_at=v_checkpoint_at,
        lease_until=v_checkpoint_at+interval '60 seconds',updated_at=v_checkpoint_at
    WHERE job_id=p_job_id;
    lease_until:=v_checkpoint_at+interval '60 seconds';
    cancellation_requested:=v_job.cancel_requested_at IS NOT NULL;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_unknown_cleanup_checkpoint_v1(
        uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,timestamptz)
FROM PUBLIC;
