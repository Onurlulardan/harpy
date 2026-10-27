-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_effect_owner_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_effect_id uuid,
    p_observed_at timestamptz
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_max integer;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance effect owner is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 5 OR
       pg_catalog.isfinite(p_observed_at) IS DISTINCT FROM true OR
       p_observed_at>pg_catalog.clock_timestamp()+interval '5 seconds' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export maintenance effect owner input is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs job
    WHERE job.id=p_job_id FOR UPDATE;
    v_max:=CASE v_job.kind WHEN 'analytics.dashboard.export.delete' THEN 3
        WHEN 'analytics.dashboard.export.expire' THEN 8 END;
    IF v_job.id IS NULL OR v_job.tenant_id IS DISTINCT FROM p_tenant_id OR
       v_job.scope_kind IS DISTINCT FROM 'tenant' OR
       v_job.scope_id IS DISTINCT FROM p_tenant_id OR v_max IS NULL OR
       v_job.authority_domain IS DISTINCT FROM 'analytics' OR
       v_job.writer_origin IS DISTINCT FROM 'v2' OR
       v_job.payload_revision IS DISTINCT FROM 1 OR
       v_job.effect_class IS DISTINCT FROM 'external' OR
       v_job.max_attempts IS DISTINCT FROM v_max OR
       v_job.payload_ref IS DISTINCT FROM p_export_id::text OR
       v_job.payload_sha256 IS DISTINCT FROM pg_catalog.encode(pg_catalog.sha256(
           pg_catalog.convert_to(p_export_id::text,'UTF8')),'hex') OR
       v_job.dedupe_key IS DISTINCT FROM v_job.kind||':'||p_export_id::text OR
       v_job.effect_idempotency_key IS DISTINCT FROM v_job.dedupe_key THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance job evidence differs';
    END IF;
    IF v_job.state IS DISTINCT FROM 'running' OR v_job.completed_at IS NOT NULL OR
       v_job.attempts NOT BETWEEN 1 AND v_max OR v_job.lease_generation<1 OR
       pg_catalog.uuid_extract_version(v_job.lease_owner_id) IS DISTINCT FROM 7 OR
       v_job.lease_until<=p_observed_at OR v_job.effect_started_at IS NULL OR
       p_observed_at<v_job.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export maintenance effect lease is stale';
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts attempt
    WHERE attempt.job_id=v_job.id AND attempt.scope_kind='tenant'
      AND attempt.scope_id=p_tenant_id AND attempt.attempt=v_job.attempts
      AND attempt.lease_generation=v_job.lease_generation FOR UPDATE;
    IF v_attempt.attempt_id IS NULL OR v_attempt.finished_at IS NOT NULL OR
       v_attempt.outcome IS NOT NULL OR
       v_attempt.lease_owner_id IS DISTINCT FROM v_job.lease_owner_id OR
       v_attempt.effect_started_at IS DISTINCT FROM v_job.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export maintenance attempt lease is stale';
    END IF;
    IF v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       pg_catalog.uuid_extract_version(v_attempt.effect_id) IS DISTINCT FROM 5 OR
       v_attempt.effect_request_sha256!~'^[0-9a-f]{64}$' OR
       v_attempt.effect_request_sha256=pg_catalog.repeat('0',64) OR
       v_attempt.effect_receipt_id IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance effect evidence differs';
    END IF;
    PERFORM 1 FROM platform.cutover_groups group_row
    JOIN platform.domain_cutovers domain_row
      ON domain_row.scope_kind=group_row.scope_kind
     AND domain_row.scope_id=group_row.scope_id
    JOIN platform.cutover_group_members member
      ON member.scope_kind=group_row.scope_kind AND member.scope_id=group_row.scope_id
     AND member.group_id=group_row.group_id AND member.domain=domain_row.domain
    JOIN platform.tenants tenant ON tenant.id=v_job.tenant_id
    WHERE group_row.scope_kind='tenant' AND group_row.scope_id=v_job.tenant_id
      AND group_row.group_id=v_job.authority_group_id
      AND group_row.generation=v_job.authority_group_generation
      AND group_row.state='v2' AND group_row.writer='v2'
      AND domain_row.domain='analytics'
      AND domain_row.generation=v_job.cutover_generation
      AND domain_row.state='v2' AND domain_row.writer='v2'
      AND tenant.deleted_at IS NULL
    FOR SHARE OF group_row,domain_row,member,tenant;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance effect owner is not current';
    END IF;
    RETURN v_job.kind;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_maintenance_effect_owner_v1(
        uuid,uuid,uuid,uuid,timestamptz) FROM PUBLIC;
