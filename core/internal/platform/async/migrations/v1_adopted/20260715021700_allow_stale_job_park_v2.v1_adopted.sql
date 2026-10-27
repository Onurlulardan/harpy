-- +goose Up
CREATE OR REPLACE TRIGGER jobs_transition_guard
BEFORE UPDATE ON platform.jobs
FOR EACH ROW WHEN (OLD.parked_at IS NOT NULL OR NEW.parked_at IS NULL)
EXECUTE FUNCTION platform.platform_async_guard_job_transition_v1();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard;

-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_stale_job_park_v2()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
DECLARE
    v_changed jsonb;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'jobs' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       OLD.state NOT IN ('queued', 'running') OR NEW.state <> 'queued' OR
       OLD.parked_at IS NOT NULL OR NEW.parked_at IS NULL OR
       OLD.effect_started_at IS NOT NULL OR
       NEW.parked_at <> NEW.updated_at OR NEW.updated_at <= OLD.updated_at OR
       NEW.parked_code <> 'JOB_AUTHORITY_STALE' OR
       NEW.parked_current_group_generation < OLD.authority_group_generation OR
       NEW.parked_current_cutover_generation < OLD.cutover_generation OR
       (NEW.parked_current_group_generation = OLD.authority_group_generation AND
        NEW.parked_current_cutover_generation = OLD.cutover_generation) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5636',
            MESSAGE = 'stale job park transition is invalid';
    END IF;

    PERFORM platform.platform_async_assert_current_authority_v1(
        OLD.scope_kind, OLD.scope_id, OLD.authority_group_id,
        NEW.parked_current_group_generation, OLD.authority_domain,
        NEW.parked_current_cutover_generation
    );

    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key)
    INTO v_changed
    FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING (key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    IF NOT (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
        'state','lease_owner_id','lease_heartbeat_at','lease_until','parked_at',
        'parked_code','parked_current_group_generation',
        'parked_current_cutover_generation','updated_at'
    ]::text[])) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5637',
            MESSAGE = 'stale job park changed immutable state';
    END IF;

    IF (OLD.state = 'queued' AND NOT EXISTS (
        SELECT 1 FROM platform.dead_letters AS dead_letter
        WHERE dead_letter.record_kind = 'root'
          AND dead_letter.source_kind = 'job'
          AND dead_letter.source_id = OLD.id
          AND dead_letter.reason_class = 'stale_generation'
          AND dead_letter.reason_code = 'JOB_AUTHORITY_STALE'
          AND dead_letter.authority_group_generation =
              NEW.parked_current_group_generation
          AND dead_letter.cutover_generation =
              NEW.parked_current_cutover_generation
          AND dead_letter.recorded_at = NEW.updated_at
    )) OR (OLD.state = 'running' AND NOT EXISTS (
        SELECT 1 FROM platform.job_attempts AS attempt
        WHERE attempt.job_id = OLD.id AND attempt.attempt = OLD.attempts
          AND attempt.lease_generation = OLD.lease_generation
          AND attempt.lease_owner_id = OLD.lease_owner_id
          AND attempt.outcome = 'lease_expired'
          AND attempt.error_class = 'conflict'
          AND attempt.error_code = 'JOB_AUTHORITY_STALE'
          AND attempt.finished_at IS NOT NULL
    )) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5638',
            MESSAGE = 'stale job park lacks durable evidence';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_stale_job_park_v2()
FROM PUBLIC;

CREATE TRIGGER jobs_stale_park_guard
BEFORE UPDATE ON platform.jobs
FOR EACH ROW WHEN (OLD.parked_at IS NULL AND NEW.parked_at IS NOT NULL)
EXECUTE FUNCTION platform.platform_async_guard_stale_job_park_v2();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_stale_park_guard;
