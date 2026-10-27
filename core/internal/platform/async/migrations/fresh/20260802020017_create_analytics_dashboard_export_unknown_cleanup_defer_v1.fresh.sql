-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_unknown_cleanup_defer_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_attempt_generation bigint,p_original_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_cleanup_generation bigint,
    p_cleanup_lease_owner_id uuid,p_observed_at timestamptz
)
RETURNS timestamptz
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_claim platform.analytics_dashboard_export_unknown_cleanup_claims%ROWTYPE;
    v_request text;
    v_deferred_at timestamptz;
    v_retry_at timestamptz;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export cleanup defer authority is invalid';
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
            MESSAGE='Analytics dashboard export cleanup defer input is invalid';
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
    IF v_job.id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export cleanup defer evidence differs';
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
       v_claim.lease_owner_id<>p_cleanup_lease_owner_id OR EXISTS (
        SELECT 1 FROM platform.analytics_dashboard_export_unknown_resolutions resolution
        WHERE resolution.job_id=p_job_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export cleanup defer lease is stale';
    END IF;
    IF v_claim.lease_until<=pg_catalog.clock_timestamp() AND
       v_claim.run_at>pg_catalog.clock_timestamp() THEN
        RETURN v_claim.run_at;
    END IF;
    IF v_claim.lease_until<=pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export cleanup defer lease expired';
    END IF;
    v_deferred_at:=GREATEST(pg_catalog.clock_timestamp(),p_observed_at,
        v_claim.updated_at+interval '1 microsecond');
    v_retry_at:=v_deferred_at+LEAST(
        interval '30 seconds'*pg_catalog.power(2::double precision,
            LEAST((p_cleanup_generation-1)::double precision,5::double precision)),
        interval '15 minutes');
    UPDATE platform.analytics_dashboard_export_unknown_cleanup_claims
    SET lease_until=v_deferred_at,run_at=v_retry_at,updated_at=v_deferred_at
    WHERE job_id=p_job_id;
    RETURN v_retry_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_unknown_cleanup_defer_v1(
        uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,timestamptz)
FROM PUBLIC;
