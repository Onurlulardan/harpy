-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_park_stale_work_v1(
    p_job_id uuid,
    p_lease_owner_id uuid,
    p_lease_generation bigint,
    p_dead_letter_id uuid,
    p_payload_canonical bytea,
    p_payload_sha256 text,
    p_redacted_envelope jsonb,
    p_error_sha256 text
) RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_group_generation bigint;
    v_domain_generation bigint;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF SESSION_USER OPERATOR(pg_catalog.<>) 'forgecrud_v2_runtime' OR
       p_job_id IS NULL OR p_lease_owner_id IS NULL OR
       p_lease_generation OPERATOR(pg_catalog.<) 1 OR p_dead_letter_id IS NULL OR
       p_payload_canonical IS NULL OR p_payload_sha256 !~ '^[0-9a-f]{64}$' OR
       p_redacted_envelope IS NULL OR p_error_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5630',
            MESSAGE = 'stale async work park input is invalid';
    END IF;

    SELECT * INTO v_job
    FROM platform.jobs
    WHERE id OPERATOR(pg_catalog.=) p_job_id;
    IF NOT FOUND THEN
        RETURN false;
    END IF;

    SELECT generation INTO v_group_generation
    FROM platform.cutover_groups
    WHERE scope_kind OPERATOR(pg_catalog.=) v_job.scope_kind
      AND scope_id OPERATOR(pg_catalog.=) v_job.scope_id
      AND group_id OPERATOR(pg_catalog.=) v_job.authority_group_id
      AND state OPERATOR(pg_catalog.=) 'v2'
      AND writer OPERATOR(pg_catalog.=) 'v2'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5631',
            MESSAGE = 'current async group authority is unavailable';
    END IF;

    SELECT generation INTO v_domain_generation
    FROM platform.domain_cutovers
    WHERE scope_kind OPERATOR(pg_catalog.=) v_job.scope_kind
      AND scope_id OPERATOR(pg_catalog.=) v_job.scope_id
      AND domain OPERATOR(pg_catalog.=) v_job.authority_domain
      AND state OPERATOR(pg_catalog.=) 'v2'
      AND writer OPERATOR(pg_catalog.=) 'v2'
    FOR SHARE;
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1 FROM platform.cutover_group_members
        WHERE scope_kind OPERATOR(pg_catalog.=) v_job.scope_kind
          AND scope_id OPERATOR(pg_catalog.=) v_job.scope_id
          AND group_id OPERATOR(pg_catalog.=) v_job.authority_group_id
          AND domain OPERATOR(pg_catalog.=) v_job.authority_domain
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5632',
            MESSAGE = 'current async domain authority is unavailable';
    END IF;

    SELECT * INTO v_job
    FROM platform.jobs
    WHERE id OPERATOR(pg_catalog.=) p_job_id
    FOR UPDATE;
    IF v_job.state OPERATOR(pg_catalog.<>) 'running' OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_generation IS DISTINCT FROM p_lease_generation OR
       v_job.lease_until IS NULL OR
       v_job.lease_until OPERATOR(pg_catalog.<=) v_now OR
       v_job.effect_started_at IS NOT NULL THEN
        RETURN false;
    END IF;
    IF v_job.payload_sha256 OPERATOR(pg_catalog.<>) p_payload_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5635',
            MESSAGE = 'stale async work payload digest differs';
    END IF;
    IF v_group_generation OPERATOR(pg_catalog.<) v_job.authority_group_generation OR
       v_domain_generation OPERATOR(pg_catalog.<) v_job.cutover_generation OR
       (v_group_generation OPERATOR(pg_catalog.=) v_job.authority_group_generation AND
        v_domain_generation OPERATOR(pg_catalog.=) v_job.cutover_generation) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5633',
            MESSAGE = 'async work authority is not stale';
    END IF;

    UPDATE platform.job_attempts
    SET finished_at = v_now,
        outcome = 'lease_expired',
        error_class = 'conflict',
        error_code = 'JOB_AUTHORITY_STALE',
        redacted_error_sha256 = p_error_sha256
    WHERE job_id OPERATOR(pg_catalog.=) v_job.id
      AND attempt OPERATOR(pg_catalog.=) v_job.attempts
      AND lease_generation OPERATOR(pg_catalog.=) p_lease_generation
      AND lease_owner_id OPERATOR(pg_catalog.=) p_lease_owner_id
      AND finished_at IS NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5634',
            MESSAGE = 'active async attempt is unavailable';
    END IF;

    UPDATE platform.jobs
    SET state = 'queued',
        lease_owner_id = NULL,
        lease_heartbeat_at = NULL,
        lease_until = NULL,
        parked_at = v_now,
        parked_code = 'JOB_AUTHORITY_STALE',
        parked_current_group_generation = v_group_generation,
        parked_current_cutover_generation = v_domain_generation,
        updated_at = v_now
    WHERE id OPERATOR(pg_catalog.=) v_job.id;

    INSERT INTO platform.dead_letters (
        id, root_id, entry_sequence, record_kind,
        tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, cutover_generation, writer_origin,
        source_kind, source_id, source_contract, source_version, source_sha256,
        reason_class, reason_code, redacted_error_sha256,
        payload_canonical, payload_sha256, redacted_envelope,
        full_payload_expires_at, metadata_expires_at, recorded_at
    ) VALUES (
        p_dead_letter_id, p_dead_letter_id, 0, 'root',
        v_job.tenant_id, v_job.scope_kind, v_job.scope_id,
        v_job.authority_group_id, v_group_generation,
        v_job.authority_domain, v_domain_generation, 'v2',
        'job', v_job.id, v_job.kind, v_job.payload_revision, v_job.payload_sha256,
        'stale_generation', 'JOB_AUTHORITY_STALE', p_error_sha256,
        p_payload_canonical, p_payload_sha256, p_redacted_envelope,
        v_now + interval '30 days', v_now + interval '365 days', v_now
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_park_stale_work_v1(
    uuid, uuid, bigint, uuid, bytea, text, jsonb, text
) FROM PUBLIC;
