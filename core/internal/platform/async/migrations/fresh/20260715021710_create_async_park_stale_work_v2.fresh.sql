-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_park_stale_work_v2(
    p_job_id uuid,
    p_expected_state text,
    p_lease_owner_id uuid,
    p_lease_generation bigint,
    p_dead_letter_id uuid
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
    v_payload bytea;
    v_payload_sha256 text;
    v_error_sha256 text;
    v_redacted_envelope jsonb;
BEGIN
    IF SESSION_USER OPERATOR(pg_catalog.<>) 'forgecrud_v2_runtime' OR
       p_job_id IS NULL OR p_dead_letter_id IS NULL OR
       NOT ((p_expected_state = 'queued' AND p_lease_owner_id IS NULL AND
             p_lease_generation >= 0) OR
            (p_expected_state = 'running' AND p_lease_owner_id IS NOT NULL AND
             p_lease_generation >= 1)) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5630',
            MESSAGE = 'stale async work park input is invalid';
    END IF;

    SELECT * INTO v_job FROM platform.jobs WHERE id = p_job_id;
    IF NOT FOUND THEN
        RETURN false;
    END IF;

    SELECT generation INTO v_group_generation
    FROM platform.cutover_groups
    WHERE scope_kind = v_job.scope_kind AND scope_id = v_job.scope_id
      AND group_id = v_job.authority_group_id
      AND state = 'v2' AND writer = 'v2'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5631',
            MESSAGE = 'current async group authority is unavailable';
    END IF;

    SELECT generation INTO v_domain_generation
    FROM platform.domain_cutovers
    WHERE scope_kind = v_job.scope_kind AND scope_id = v_job.scope_id
      AND domain = v_job.authority_domain
      AND state = 'v2' AND writer = 'v2'
    FOR SHARE;
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1 FROM platform.cutover_group_members
        WHERE scope_kind = v_job.scope_kind AND scope_id = v_job.scope_id
          AND group_id = v_job.authority_group_id
          AND domain = v_job.authority_domain
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5632',
            MESSAGE = 'current async domain authority is unavailable';
    END IF;

    SELECT * INTO v_job FROM platform.jobs WHERE id = p_job_id FOR UPDATE;
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    IF v_job.state IS DISTINCT FROM p_expected_state OR
       v_job.parked_at IS NOT NULL OR v_job.effect_started_at IS NOT NULL OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_generation IS DISTINCT FROM p_lease_generation THEN
        RETURN false;
    END IF;
    IF v_group_generation < v_job.authority_group_generation OR
       v_domain_generation < v_job.cutover_generation OR
       (v_group_generation = v_job.authority_group_generation AND
        v_domain_generation = v_job.cutover_generation) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5633',
            MESSAGE = 'async work authority is not stale';
    END IF;

    v_payload := pg_catalog.convert_to(
        '{"contract_version":"forgecrud.job-stale-reference.v1","job_id":' ||
        pg_catalog.to_json(v_job.id::text)::text || ',"kind":' ||
        pg_catalog.to_json(v_job.kind)::text || ',"payload_reference":' ||
        pg_catalog.to_json(v_job.payload_ref)::text || ',"payload_revision":' ||
        v_job.payload_revision::text || ',"payload_sha256":' ||
        pg_catalog.to_json(v_job.payload_sha256)::text || '}', 'UTF8'
    );
    v_payload_sha256 := pg_catalog.encode(pg_catalog.sha256(v_payload), 'hex');
    v_error_sha256 := pg_catalog.encode(
        pg_catalog.sha256(pg_catalog.convert_to('JOB_AUTHORITY_STALE', 'UTF8')), 'hex'
    );
    v_redacted_envelope := pg_catalog.jsonb_build_object(
        'contract_version', 'forgecrud.job-stale-reference.v1',
        'job_id', v_job.id, 'kind', v_job.kind,
        'payload_revision', v_job.payload_revision,
        'payload_sha256', v_job.payload_sha256
    );

    INSERT INTO platform.dead_letters (
        id, root_id, entry_sequence, record_kind, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation, authority_domain,
        cutover_generation, writer_origin, source_kind, source_id, source_contract,
        source_version, source_sha256, reason_class, reason_code,
        redacted_error_sha256, payload_canonical, payload_sha256,
        redacted_envelope, full_payload_expires_at, metadata_expires_at, recorded_at
    ) VALUES (
        p_dead_letter_id, p_dead_letter_id, 0, 'root', v_job.tenant_id,
        v_job.scope_kind, v_job.scope_id, v_job.authority_group_id,
        v_group_generation, v_job.authority_domain, v_domain_generation, 'v2',
        'job', v_job.id, v_job.kind, v_job.payload_revision, v_job.payload_sha256,
        'stale_generation', 'JOB_AUTHORITY_STALE', v_error_sha256,
        v_payload, v_payload_sha256, v_redacted_envelope,
        v_now + interval '30 days', v_now + interval '365 days', v_now
    );

    IF v_job.state = 'running' THEN
        UPDATE platform.job_attempts
        SET finished_at = v_now, outcome = 'lease_expired',
            error_class = 'conflict', error_code = 'JOB_AUTHORITY_STALE',
            redacted_error_sha256 = v_error_sha256
        WHERE job_id = v_job.id AND attempt = v_job.attempts
          AND lease_generation = v_job.lease_generation
          AND lease_owner_id = v_job.lease_owner_id AND finished_at IS NULL;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = 'P5634',
                MESSAGE = 'active async attempt is unavailable';
        END IF;
    END IF;

    UPDATE platform.jobs
    SET state = 'queued', lease_owner_id = NULL, lease_heartbeat_at = NULL,
        lease_until = NULL, parked_at = v_now, parked_code = 'JOB_AUTHORITY_STALE',
        parked_current_group_generation = v_group_generation,
        parked_current_cutover_generation = v_domain_generation, updated_at = v_now
    WHERE id = v_job.id;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_park_stale_work_v2(
    uuid,text,uuid,bigint,uuid
) FROM PUBLIC;
