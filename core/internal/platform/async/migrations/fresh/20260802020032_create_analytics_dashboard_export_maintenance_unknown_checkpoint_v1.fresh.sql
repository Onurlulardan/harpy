-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_checkpoint_v1(
    p_kind text,p_tenant_id uuid,p_tenant_schema name,p_export_id uuid,
    p_authority_group_id text,p_authority_group_generation bigint,
    p_cutover_generation bigint,p_job_id uuid,p_attempt_id uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_unknown_at timestamptz,p_cleanup_generation bigint,
    p_cleanup_lease_owner_id uuid,p_observed_at timestamptz
)
RETURNS timestamptz
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_claim platform.analytics_dashboard_export_maintenance_unknown_claims%ROWTYPE;
    v_schema name;
    v_request text;
    v_checkpoint_at timestamptz;
    v_max integer;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance checkpoint authority is invalid';
    END IF;
    v_max:=CASE p_kind
        WHEN 'analytics.dashboard.export.delete' THEN 3
        WHEN 'analytics.dashboard.export.expire' THEN 8 END;
    IF v_max IS NULL OR p_tenant_id IS NULL OR
       p_tenant_schema::text!~'^[a-z][a-z0-9_]{0,62}$' OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_authority_group_id!~'^[a-z][a-z0-9_]{0,62}$' OR
       p_authority_group_generation<1 OR p_cutover_generation<1 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_number NOT BETWEEN 1 AND v_max OR p_lease_generation<1 OR
       pg_catalog.uuid_extract_version(p_original_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 5 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_request_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       NOT pg_catalog.isfinite(p_unknown_at) OR p_unknown_at>p_observed_at OR
       p_cleanup_generation<1 OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       NOT pg_catalog.isfinite(p_observed_at) OR
       p_observed_at>pg_catalog.clock_timestamp()+interval '5 seconds' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export maintenance checkpoint input is invalid';
    END IF;
    v_request:=pg_catalog.encode(p_request_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_group_id=p_authority_group_id
      AND job.authority_group_generation=p_authority_group_generation
      AND job.authority_domain='analytics'
      AND job.cutover_generation=p_cutover_generation AND job.writer_origin='v2'
      AND job.kind=p_kind AND job.payload_revision=1
      AND job.effect_class='external' AND job.max_attempts=v_max
      AND job.payload_ref=p_export_id::text AND job.state='dead'
      AND job.last_error_class IS NULL
      AND job.last_error_code='JOB_OUTCOME_UNKNOWN'
      AND job.completed_at=p_unknown_at FOR SHARE;
    IF v_job.id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance checkpoint evidence differs';
    END IF;
    SELECT tenant.schema_name INTO v_schema FROM platform.tenants AS tenant
    WHERE tenant.id=p_tenant_id AND tenant.deleted_at IS NULL FOR SHARE;
    IF v_schema IS DISTINCT FROM p_tenant_schema THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export maintenance tenant schema changed';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    SELECT claim.* INTO v_claim
    FROM platform.analytics_dashboard_export_maintenance_unknown_claims AS claim
    WHERE claim.job_id=p_job_id AND claim.kind=p_kind
      AND claim.tenant_id=p_tenant_id AND claim.tenant_schema=p_tenant_schema
      AND claim.export_id=p_export_id AND claim.attempt_id=p_attempt_id
      AND claim.authority_group_id=p_authority_group_id
      AND claim.authority_group_generation=p_authority_group_generation
      AND claim.cutover_generation=p_cutover_generation
      AND claim.attempt_number=p_attempt_number
      AND claim.lease_generation=p_lease_generation
      AND claim.original_lease_owner_id=p_original_lease_owner_id
      AND claim.effect_id=p_effect_id AND claim.request_sha256=v_request
      AND claim.unknown_at=p_unknown_at FOR UPDATE;
    IF v_claim.job_id IS NULL OR
       v_claim.claim_generation<>p_cleanup_generation OR
       v_claim.cleanup_lease_owner_id<>p_cleanup_lease_owner_id OR
       v_claim.lease_until<=pg_catalog.clock_timestamp() OR EXISTS (
        SELECT 1
        FROM platform.analytics_dashboard_export_maintenance_unknown_resolutions resolution
        WHERE resolution.job_id=p_job_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export maintenance checkpoint lease is stale';
    END IF;
    v_checkpoint_at:=GREATEST(pg_catalog.clock_timestamp(),p_observed_at,
        v_claim.updated_at+interval '1 microsecond');
    UPDATE platform.analytics_dashboard_export_maintenance_unknown_claims
    SET lease_heartbeat_at=v_checkpoint_at,
        lease_until=v_checkpoint_at+interval '60 seconds',updated_at=v_checkpoint_at
    WHERE job_id=p_job_id;
    RETURN v_checkpoint_at+interval '60 seconds';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_maintenance_unknown_checkpoint_v1(
        text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,uuid,uuid,
        bytea,timestamptz,bigint,uuid,timestamptz)
FROM PUBLIC;
