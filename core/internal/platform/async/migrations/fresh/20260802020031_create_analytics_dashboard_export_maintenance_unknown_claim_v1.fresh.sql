-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_unknown_claim_v1(
    p_kind text,p_tenant_id uuid,p_tenant_schema name,p_export_id uuid,
    p_authority_group_id text,p_authority_group_generation bigint,
    p_cutover_generation bigint,p_job_id uuid,p_attempt_id uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_unknown_at timestamptz,p_cleanup_lease_owner_id uuid,
    p_observed_at timestamptz
)
RETURNS TABLE(claim_state text,cleanup_generation bigint,lease_until timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_claim platform.analytics_dashboard_export_maintenance_unknown_claims%ROWTYPE;
    v_resolution platform.analytics_dashboard_export_maintenance_unknown_resolutions%ROWTYPE;
    v_schema name;
    v_request text;
    v_claimed_at timestamptz;
    v_max integer;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance claim authority is invalid';
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
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       NOT pg_catalog.isfinite(p_observed_at) OR
       p_observed_at>pg_catalog.clock_timestamp()+interval '5 seconds' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export maintenance claim input is invalid';
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
      AND job.payload_ref=p_export_id::text
      AND job.payload_sha256=pg_catalog.encode(pg_catalog.sha256(
          pg_catalog.convert_to(p_export_id::text,'UTF8')),'hex')
      AND job.dedupe_key=p_kind||':'||p_export_id::text
      AND job.effect_idempotency_key=job.dedupe_key FOR UPDATE;
    IF v_job.id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance job evidence differs';
    END IF;
    SELECT tenant.schema_name INTO v_schema FROM platform.tenants AS tenant
    WHERE tenant.id=p_tenant_id AND tenant.deleted_at IS NULL FOR SHARE;
    IF v_schema IS DISTINCT FROM p_tenant_schema THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export maintenance tenant schema changed';
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
       AND attempt.attempt=p_attempt_number
       AND attempt.lease_generation=p_lease_generation
       AND attempt.lease_owner_id=p_original_lease_owner_id FOR SHARE;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    SELECT resolution.* INTO v_resolution
    FROM platform.analytics_dashboard_export_maintenance_unknown_resolutions resolution
    WHERE resolution.job_id=p_job_id FOR SHARE;
    IF FOUND THEN
        IF v_resolution.kind<>p_kind OR v_resolution.tenant_id<>p_tenant_id OR
           v_resolution.tenant_schema<>p_tenant_schema OR
           v_resolution.export_id<>p_export_id OR
           v_resolution.authority_group_id<>p_authority_group_id OR
           v_resolution.authority_group_generation<>p_authority_group_generation OR
           v_resolution.cutover_generation<>p_cutover_generation OR
           v_resolution.attempt_id<>p_attempt_id OR
           v_resolution.attempt_number<>p_attempt_number OR
           v_resolution.lease_generation<>p_lease_generation OR
           v_resolution.original_lease_owner_id<>p_original_lease_owner_id OR
           v_resolution.effect_id<>p_effect_id OR
           v_resolution.request_sha256<>v_request OR
           v_resolution.original_unknown_at<>p_unknown_at OR
           v_job.state<>'succeeded' OR v_attempt.outcome<>'succeeded' OR
           v_attempt.effect_receipt_id<>v_resolution.effect_receipt_id THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export maintenance resolution differs';
        END IF;
        claim_state:='resolved'; RETURN NEXT; RETURN;
    END IF;
    IF NOT COALESCE((v_job.state='dead' AND
       v_job.last_error_class IS NULL AND
       v_job.last_error_code='JOB_OUTCOME_UNKNOWN' AND
       v_job.last_error_sha256 IS NULL AND v_job.completed_at=p_unknown_at AND
       v_job.updated_at=v_job.completed_at AND
       v_job.retain_until=v_job.completed_at+interval '365 days' AND
       v_job.lease_owner_id IS NULL AND v_job.lease_heartbeat_at IS NULL AND
       v_job.lease_until IS NULL AND v_job.attempts=p_attempt_number AND
       v_job.lease_generation=p_lease_generation AND
       v_attempt.finished_at=v_job.completed_at AND
       v_attempt.outcome='dead' AND v_attempt.error_class IS NULL AND
       v_attempt.error_code='JOB_OUTCOME_UNKNOWN' AND
       v_attempt.redacted_error_sha256 IS NULL AND
       v_attempt.effect_started_at IS NOT NULL AND
       v_attempt.effect_id=p_effect_id AND
       v_attempt.effect_request_sha256=v_request AND
       v_attempt.effect_receipt_id IS NULL),false) OR EXISTS (
        SELECT 1 FROM platform.idempotency_receipts AS receipt
        WHERE receipt.record_kind='effect' AND receipt.scope_kind=v_job.scope_kind
          AND receipt.scope_id=v_job.scope_id
          AND receipt.authority_group_id=v_job.authority_group_id
          AND receipt.authority_group_generation=v_job.authority_group_generation
          AND receipt.authority_domain=v_job.authority_domain
          AND receipt.cutover_generation=v_job.cutover_generation
          AND receipt.effect_key=v_job.effect_idempotency_key
          AND receipt.effect_id=p_effect_id AND receipt.request_sha256=v_request
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance unknown evidence differs';
    END IF;
    SELECT claim.* INTO v_claim
    FROM platform.analytics_dashboard_export_maintenance_unknown_claims AS claim
    WHERE claim.job_id=p_job_id FOR UPDATE;
    IF FOUND AND (v_claim.kind<>p_kind OR v_claim.tenant_id<>p_tenant_id OR
       v_claim.tenant_schema<>p_tenant_schema OR v_claim.export_id<>p_export_id OR
       v_claim.authority_group_id<>p_authority_group_id OR
       v_claim.authority_group_generation<>p_authority_group_generation OR
       v_claim.cutover_generation<>p_cutover_generation OR
       v_claim.attempt_id<>p_attempt_id OR
       v_claim.attempt_number<>p_attempt_number OR
       v_claim.lease_generation<>p_lease_generation OR
       v_claim.original_lease_owner_id<>p_original_lease_owner_id OR
       v_claim.effect_id<>p_effect_id OR v_claim.request_sha256<>v_request OR
       v_claim.unknown_at<>p_unknown_at) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance claim replay differs';
    END IF;
    IF FOUND AND v_claim.lease_until>pg_catalog.clock_timestamp() THEN
        IF v_claim.cleanup_lease_owner_id=p_cleanup_lease_owner_id THEN
            claim_state:='claimed'; cleanup_generation:=v_claim.claim_generation;
            lease_until:=v_claim.lease_until; RETURN NEXT; RETURN;
        END IF;
        claim_state:='busy'; RETURN NEXT; RETURN;
    END IF;
    IF FOUND AND v_claim.run_at>pg_catalog.clock_timestamp() THEN
        claim_state:='busy'; RETURN NEXT; RETURN;
    END IF;
    v_claimed_at:=GREATEST(pg_catalog.clock_timestamp(),p_observed_at,
        COALESCE(v_claim.updated_at,p_observed_at)+interval '1 microsecond');
    IF FOUND THEN
        UPDATE platform.analytics_dashboard_export_maintenance_unknown_claims
        SET claim_generation=claim_generation+1,
            cleanup_lease_owner_id=p_cleanup_lease_owner_id,
            lease_heartbeat_at=v_claimed_at,
            lease_until=v_claimed_at+interval '60 seconds',updated_at=v_claimed_at
        WHERE job_id=p_job_id RETURNING claim_generation INTO cleanup_generation;
    ELSE
        cleanup_generation:=1;
        INSERT INTO platform.analytics_dashboard_export_maintenance_unknown_claims(
            job_id,kind,tenant_id,tenant_schema,export_id,authority_group_id,
            authority_group_generation,cutover_generation,attempt_id,attempt_number,
            lease_generation,original_lease_owner_id,effect_id,request_sha256,
            unknown_at,claim_generation,cleanup_lease_owner_id,lease_heartbeat_at,
            lease_until,run_at,created_at,updated_at
        ) VALUES (p_job_id,p_kind,p_tenant_id,p_tenant_schema,p_export_id,
            v_job.authority_group_id,v_job.authority_group_generation,
            v_job.cutover_generation,p_attempt_id,p_attempt_number,p_lease_generation,
            p_original_lease_owner_id,p_effect_id,v_request,v_job.completed_at,1,
            p_cleanup_lease_owner_id,v_claimed_at,v_claimed_at+interval '60 seconds',
            v_claimed_at,v_claimed_at,v_claimed_at);
    END IF;
    claim_state:='claimed'; lease_until:=v_claimed_at+interval '60 seconds';
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_maintenance_unknown_claim_v1(
        text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,uuid,uuid,
        bytea,timestamptz,uuid,timestamptz)
FROM PUBLIC;
