-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_unknown_evidence_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_attempt_generation bigint,p_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea
)
RETURNS timestamptz
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_request text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export unknown evidence authority is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_request_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export unknown evidence input is invalid';
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
      AND job.last_error_code='JOB_OUTCOME_UNKNOWN'
      AND job.last_error_sha256 IS NULL FOR SHARE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id
      AND attempt.finished_at=v_job.completed_at AND attempt.outcome='dead'
      AND attempt.error_class IS NULL
      AND attempt.error_code='JOB_OUTCOME_UNKNOWN'
      AND attempt.redacted_error_sha256 IS NULL
      AND attempt.effect_id=p_effect_id
      AND attempt.effect_request_sha256=v_request
      AND attempt.effect_started_at IS NOT NULL
      AND attempt.effect_receipt_id IS NULL FOR SHARE;
    IF v_job.id IS NULL OR v_attempt.attempt_id IS NULL OR EXISTS (
        SELECT 1 FROM platform.analytics_dashboard_export_unknown_resolutions resolution
        WHERE resolution.job_id=p_job_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown evidence differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    RETURN v_job.completed_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_unknown_evidence_v1(
        uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea)
FROM PUBLIC;
