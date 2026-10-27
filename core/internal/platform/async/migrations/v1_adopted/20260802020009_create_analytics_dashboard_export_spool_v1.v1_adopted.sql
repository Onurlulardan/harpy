-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_spool_state_v1(
    p_tenant_id uuid,p_export_id uuid,p_attempt_generation bigint,
    p_lease_owner_id uuid
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export spool state authority is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.tenant_id=p_tenant_id AND job.scope_kind='tenant'
      AND job.scope_id=p_tenant_id AND job.authority_domain='analytics'
      AND job.writer_origin='v2'
      AND job.kind='analytics.dashboard.export.render'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=3 AND job.payload_ref=p_export_id::text
      AND job.payload_sha256=pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(p_export_id::text,'UTF8')),'hex')
      AND job.dedupe_key='analytics.dashboard.export.render:'||p_export_id::text
      AND job.effect_idempotency_key=job.dedupe_key;
    IF NOT FOUND OR v_job.attempts<>p_attempt_generation OR
       v_job.lease_generation<>p_attempt_generation THEN
        RETURN 'inactive';
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=v_job.id AND attempt.scope_kind='tenant'
      AND attempt.scope_id=p_tenant_id AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id;
    IF NOT FOUND THEN RETURN 'inactive'; END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.analytics_dashboard_export_unknown_resolutions AS resolution
        WHERE resolution.tenant_id=p_tenant_id
          AND resolution.export_id=p_export_id AND resolution.job_id=v_job.id
          AND resolution.attempt_id=v_attempt.attempt_id
          AND resolution.attempt_generation=p_attempt_generation
          AND resolution.lease_owner_id=p_lease_owner_id
          AND resolution.effect_id=v_attempt.effect_id
          AND resolution.request_sha256=v_attempt.effect_request_sha256
    ) THEN
        RETURN 'inactive';
    END IF;
    IF v_job.state='running' AND v_attempt.finished_at IS NULL AND
       v_job.lease_owner_id=p_lease_owner_id AND
       v_job.lease_until>pg_catalog.clock_timestamp() THEN
        RETURN 'running';
    END IF;
    IF v_job.state='dead' AND v_job.last_error_class IS NULL AND
       v_job.last_error_code='JOB_OUTCOME_UNKNOWN' AND
       v_attempt.outcome='dead' AND v_attempt.error_class IS NULL AND
       v_attempt.error_code='JOB_OUTCOME_UNKNOWN' AND
       v_attempt.effect_started_at IS NOT NULL AND
       v_attempt.effect_receipt_id IS NULL THEN
        RETURN 'outcome_unknown';
    END IF;
    RETURN 'inactive';
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_spool_state_v1(
        uuid,uuid,bigint,uuid)
FROM PUBLIC;

CREATE FUNCTION platform.platform_async_analytics_dashboard_export_spool_tenant_v1(
    p_export_id uuid
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_tenant_id uuid;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export spool tenant authority is invalid';
    END IF;
    SELECT job.tenant_id INTO v_tenant_id FROM platform.jobs AS job
    WHERE job.scope_kind='tenant' AND job.scope_id=job.tenant_id
      AND job.authority_domain='analytics' AND job.writer_origin='v2'
      AND job.kind='analytics.dashboard.export.render'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=3 AND job.payload_ref=p_export_id::text
      AND job.payload_sha256=pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(p_export_id::text,'UTF8')),'hex')
      AND job.dedupe_key='analytics.dashboard.export.render:'||p_export_id::text
      AND job.effect_idempotency_key=job.dedupe_key
    ORDER BY job.tenant_id LIMIT 1;
    IF EXISTS (
        SELECT 1 FROM platform.jobs AS other
        WHERE other.scope_kind='tenant' AND other.scope_id=other.tenant_id
          AND other.authority_domain='analytics' AND other.writer_origin='v2'
          AND other.kind='analytics.dashboard.export.render'
          AND other.payload_revision=1 AND other.effect_class='external'
          AND other.max_attempts=3 AND other.payload_ref=p_export_id::text
          AND other.tenant_id IS DISTINCT FROM v_tenant_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export spool tenant is ambiguous';
    END IF;
    RETURN v_tenant_id;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_spool_tenant_v1(uuid)
FROM PUBLIC;
-- +goose StatementEnd
