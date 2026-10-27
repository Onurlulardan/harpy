-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_stale_external_job_unknown_v1(
    OLD platform.jobs,
    NEW platform.jobs,
    p_current_user name,
    p_session_user name
) RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_attempt platform.job_attempts%ROWTYPE;
    v_group_generation bigint;
    v_domain_generation bigint;
    v_changed jsonb;
BEGIN
    IF pg_catalog.pg_trigger_depth() <> 1 OR
       p_current_user IS DISTINCT FROM 'forgecrud_owner' OR
       p_session_user IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       OLD.state IS DISTINCT FROM 'running' OR
       OLD.effect_class IS DISTINCT FROM 'external' OR
       OLD.effect_started_at IS NULL OR NEW.state IS DISTINCT FROM 'dead' OR
       NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at OR
       NEW.lease_owner_id IS NOT NULL OR NEW.lease_heartbeat_at IS NOT NULL OR
       NEW.lease_until IS NOT NULL OR NEW.last_error_class IS NOT NULL OR
       NEW.last_error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       NEW.last_error_sha256 IS NOT NULL OR
       NEW.completed_at IS DISTINCT FROM NEW.updated_at OR
       NEW.updated_at <= OLD.updated_at OR
       NEW.retain_until < NEW.completed_at + interval '365 days' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5674',
            MESSAGE = 'stale external job finalization is invalid';
    END IF;

    SELECT attempt.* INTO v_attempt
    FROM platform.job_attempts AS attempt
    WHERE attempt.job_id = OLD.id AND attempt.attempt = OLD.attempts
      AND attempt.lease_generation = OLD.lease_generation
      AND attempt.lease_owner_id = OLD.lease_owner_id
    FOR SHARE;
    IF NOT FOUND OR v_attempt.finished_at IS DISTINCT FROM NEW.updated_at OR
       v_attempt.outcome IS DISTINCT FROM 'dead' OR
       v_attempt.error_class IS NOT NULL OR
       v_attempt.error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       v_attempt.redacted_error_sha256 IS NOT NULL OR
       v_attempt.effect_started_at IS DISTINCT FROM OLD.effect_started_at OR
       v_attempt.effect_receipt_id IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5675',
            MESSAGE = 'stale external job lacks its terminal attempt';
    END IF;

    SELECT generation INTO v_group_generation
    FROM platform.cutover_groups
    WHERE scope_kind = OLD.scope_kind AND scope_id = OLD.scope_id
      AND group_id = OLD.authority_group_id
      AND state = 'v2' AND writer = 'v2'
    FOR SHARE;
    SELECT generation INTO v_domain_generation
    FROM platform.domain_cutovers
    WHERE scope_kind = OLD.scope_kind AND scope_id = OLD.scope_id
      AND domain = OLD.authority_domain
      AND state = 'v2' AND writer = 'v2'
    FOR SHARE;
    IF v_group_generation IS NULL OR v_domain_generation IS NULL OR
       v_group_generation < OLD.authority_group_generation OR
       v_domain_generation < OLD.cutover_generation OR
       (v_group_generation = OLD.authority_group_generation AND
        v_domain_generation = OLD.cutover_generation) OR NOT EXISTS (
        SELECT 1 FROM platform.cutover_group_members
        WHERE scope_kind = OLD.scope_kind AND scope_id = OLD.scope_id
          AND group_id = OLD.authority_group_id
          AND domain = OLD.authority_domain
    ) OR EXISTS (
        SELECT 1 FROM platform.idempotency_receipts AS receipt
        WHERE receipt.record_kind = 'effect'
          AND receipt.scope_kind = OLD.scope_kind
          AND receipt.scope_id = OLD.scope_id
          AND receipt.authority_group_id = OLD.authority_group_id
          AND receipt.authority_group_generation = OLD.authority_group_generation
          AND receipt.authority_domain = OLD.authority_domain
          AND receipt.cutover_generation = OLD.cutover_generation
          AND receipt.effect_key = OLD.effect_idempotency_key
          AND receipt.effect_id = v_attempt.effect_id
          AND receipt.request_sha256 = v_attempt.effect_request_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5676',
            MESSAGE = 'external job outcome is not unknown under stale authority';
    END IF;

    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key)
    INTO v_changed
    FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING (key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    IF v_changed IS DISTINCT FROM pg_catalog.to_jsonb(ARRAY[
        'completed_at','last_error_code','lease_heartbeat_at','lease_owner_id',
        'lease_until','retain_until','state','updated_at'
    ]::text[]) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5677',
            MESSAGE = 'stale external job changed immutable state';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_validate_stale_external_job_unknown_v1(
        platform.jobs,platform.jobs,name,name
    ) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_guard_job_transition_v2()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'jobs' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5610',
            MESSAGE = 'job transition invocation is invalid';
    END IF;
    IF CURRENT_USER = 'forgecrud_owner' AND OLD.parked_at IS NULL AND
       NEW.parked_at IS NOT NULL THEN
        NULL;
    ELSIF CURRENT_USER = 'forgecrud_owner' AND OLD.state = 'running' AND
       OLD.effect_started_at IS NOT NULL AND NEW.state = 'dead' AND
       NEW.last_error_code = 'JOB_OUTCOME_UNKNOWN' THEN
        PERFORM platform.platform_async_validate_stale_external_job_unknown_v1(
            OLD, NEW, CURRENT_USER, SESSION_USER
        );
    ELSE
        PERFORM platform.platform_async_validate_job_transition_v1(
            OLD, NEW, CURRENT_USER, SESSION_USER
        );
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_job_transition_v2() FROM PUBLIC;

CREATE OR REPLACE TRIGGER jobs_transition_guard
BEFORE UPDATE ON platform.jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_job_transition_v2();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard;
