-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_claim_v1(
    p_tenant_id uuid,
    p_export_id uuid,
    p_job_id uuid,
    p_attempt_generation bigint,
    p_lease_owner_id uuid
)
RETURNS TABLE(
    claim_state text,
    attempt_id uuid,
    claimed_at timestamptz,
    lease_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_quota platform.artifact_quota_counters%ROWTYPE;
    v_started_at timestamptz;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export claim authority is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_domain='analytics' AND job.writer_origin='v2'
      AND job.kind='analytics.dashboard.export.render'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=3 AND job.payload_ref=p_export_id::text
      AND job.payload_sha256=pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(p_export_id::text,'UTF8')),'hex')
      AND job.dedupe_key='analytics.dashboard.export.render:'||p_export_id::text
      AND job.effect_idempotency_key=job.dedupe_key
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export candidate changed';
    END IF;
    IF v_job.state='running' AND v_job.attempts=p_attempt_generation AND
       v_job.lease_generation=p_attempt_generation AND
       v_job.lease_owner_id=p_lease_owner_id THEN
        SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
        WHERE attempt.job_id=p_job_id AND attempt.attempt=p_attempt_generation
          AND attempt.lease_generation=p_attempt_generation
          AND attempt.lease_owner_id=p_lease_owner_id
          AND attempt.finished_at IS NULL;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export claim replay differs';
        END IF;
        claim_state:='claimed'; attempt_id:=v_attempt.attempt_id;
        claimed_at:=v_attempt.started_at; lease_until:=v_job.lease_until;
        RETURN NEXT; RETURN;
    END IF;
    IF v_job.state<>'queued' OR v_job.parked_at IS NOT NULL OR
       v_job.effect_started_at IS NOT NULL OR
       v_job.cancel_requested_at IS NOT NULL OR
       v_job.attempts+1<>p_attempt_generation OR
       v_job.attempts>=v_job.max_attempts OR
       v_job.run_at>pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export candidate changed';
    END IF;
    PERFORM 1 FROM platform.tenants AS tenant
    WHERE tenant.id=p_tenant_id AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE','TRIAL') FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export tenant is not claimable';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_group_generation,v_job.authority_domain,
        v_job.cutover_generation);
    SELECT quota.* INTO v_quota
    FROM platform.artifact_quota_counters AS quota
    WHERE quota.tenant_id=p_tenant_id FOR UPDATE;
    IF NOT FOUND OR v_quota.fenced THEN
        RAISE EXCEPTION USING ERRCODE='P6049',
            MESSAGE='Analytics dashboard export artifact quota is fenced';
    END IF;
    IF v_quota.used_ciphertext_bytes+v_quota.reserved_ciphertext_bytes+
       2147483648>21474836480 THEN
        claim_state:='quota_wait'; RETURN NEXT; RETURN;
    END IF;
    v_started_at:=GREATEST(
        pg_catalog.clock_timestamp(),v_job.updated_at+interval '1 microsecond');
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes+2147483648,
        revision=revision+1,updated_at=GREATEST(
            updated_at+interval '1 microsecond',v_started_at)
    WHERE tenant_id=p_tenant_id;
    UPDATE platform.jobs SET state='running',attempts=p_attempt_generation,
        lease_generation=p_attempt_generation,lease_owner_id=p_lease_owner_id,
        lease_heartbeat_at=v_started_at,
        lease_until=v_started_at+interval '60 seconds',updated_at=v_started_at
    WHERE id=p_job_id;
    attempt_id:=pg_catalog.uuidv7();
    INSERT INTO platform.job_attempts(
        attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,
        lease_owner_id,started_at
    ) VALUES (
        attempt_id,p_job_id,'tenant',p_tenant_id,p_attempt_generation,
        p_attempt_generation,p_lease_owner_id,v_started_at);
    claim_state:='claimed'; claimed_at:=v_started_at;
    lease_until:=v_started_at+interval '60 seconds';
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_claim_v1(
        uuid,uuid,uuid,bigint,uuid)
FROM PUBLIC;
