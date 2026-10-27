-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_finalize_stale_external_unknown_v1(
    p_job_id uuid,
    p_lease_owner_id uuid,
    p_lease_generation bigint,
    p_attempt_id uuid,
    p_effect_started_at timestamptz
) RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_group_generation bigint;
    v_domain_generation bigint;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_job_id IS NULL OR p_lease_owner_id IS NULL OR
       p_lease_generation < 1 OR p_attempt_id IS NULL OR
       p_effect_started_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5678',
            MESSAGE = 'stale external finalization input is invalid';
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
        RETURN false;
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
        RETURN false;
    END IF;

    SELECT * INTO v_job FROM platform.jobs WHERE id = p_job_id FOR UPDATE;
    IF NOT FOUND OR v_job.state IS DISTINCT FROM 'running' OR
       v_job.effect_class IS DISTINCT FROM 'external' OR
       v_job.effect_started_at IS DISTINCT FROM p_effect_started_at OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_generation IS DISTINCT FROM p_lease_generation OR
       v_job.parked_at IS NOT NULL OR
       v_group_generation < v_job.authority_group_generation OR
       v_domain_generation < v_job.cutover_generation OR
       (v_group_generation = v_job.authority_group_generation AND
        v_domain_generation = v_job.cutover_generation) THEN
        RETURN false;
    END IF;

    SELECT * INTO v_attempt
    FROM platform.job_attempts
    WHERE attempt_id = p_attempt_id AND job_id = v_job.id
      AND attempt = v_job.attempts
      AND lease_generation = p_lease_generation
      AND lease_owner_id = p_lease_owner_id
    FOR UPDATE;
    IF NOT FOUND OR v_attempt.finished_at IS NOT NULL OR
       v_attempt.effect_started_at IS DISTINCT FROM p_effect_started_at OR
       v_attempt.effect_id IS NULL OR v_attempt.effect_request_sha256 IS NULL OR
       EXISTS (
        SELECT 1 FROM platform.idempotency_receipts AS receipt
        WHERE receipt.record_kind = 'effect'
          AND receipt.scope_kind = v_job.scope_kind
          AND receipt.scope_id = v_job.scope_id
          AND receipt.authority_group_id = v_job.authority_group_id
          AND receipt.authority_group_generation = v_job.authority_group_generation
          AND receipt.authority_domain = v_job.authority_domain
          AND receipt.cutover_generation = v_job.cutover_generation
          AND receipt.effect_key = v_job.effect_idempotency_key
          AND receipt.effect_id = v_attempt.effect_id
          AND receipt.request_sha256 = v_attempt.effect_request_sha256
    ) THEN
        RETURN false;
    END IF;

    v_now := GREATEST(
        pg_catalog.clock_timestamp(),
        v_job.updated_at + interval '1 microsecond',
        v_attempt.effect_started_at
    );
    UPDATE platform.job_attempts
    SET finished_at = v_now, outcome = 'dead', error_class = NULL,
        error_code = 'JOB_OUTCOME_UNKNOWN', redacted_error_sha256 = NULL,
        effect_receipt_id = NULL
    WHERE attempt_id = v_attempt.attempt_id AND finished_at IS NULL;
    IF NOT FOUND THEN
        RETURN false;
    END IF;

    UPDATE platform.jobs
    SET state = 'dead', lease_owner_id = NULL, lease_heartbeat_at = NULL,
        lease_until = NULL, last_error_class = NULL,
        last_error_code = 'JOB_OUTCOME_UNKNOWN', last_error_sha256 = NULL,
        retain_until = v_now + interval '365 days',
        updated_at = v_now, completed_at = v_now
    WHERE id = v_job.id;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_finalize_stale_external_unknown_v1(
        uuid,uuid,bigint,uuid,timestamptz
    ) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_finalize_stale_external_unknown_v1(
        uuid,uuid,bigint,uuid,timestamptz
    ) TO harpy_runtime;
