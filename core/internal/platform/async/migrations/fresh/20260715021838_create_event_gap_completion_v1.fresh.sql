-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_complete_event_gap_job_v1(
    p_job platform.jobs,
    p_attempt_id uuid,
    p_lease_owner_id uuid,
    p_lease_generation bigint
)
RETURNS timestamptz
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_completed_at timestamptz;
    v_count bigint;
BEGIN
    v_completed_at := GREATEST(
        pg_catalog.clock_timestamp(),p_job.updated_at + interval '1 microsecond'
    );
    UPDATE platform.job_attempts AS attempt
    SET finished_at = v_completed_at,outcome = 'succeeded'
    WHERE attempt.attempt_id = p_attempt_id
      AND attempt.job_id = p_job.id
      AND attempt.lease_owner_id = p_lease_owner_id
      AND attempt.lease_generation = p_lease_generation
      AND attempt.finished_at IS NULL
      AND attempt.effect_started_at IS NULL
      AND attempt.effect_receipt_id IS NULL;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5695',
            MESSAGE = 'event gap attempt completion failed';
    END IF;

    UPDATE platform.jobs AS job
    SET state = 'succeeded',lease_owner_id = NULL,
        lease_heartbeat_at = NULL,lease_until = NULL,
        last_error_class = NULL,last_error_code = NULL,last_error_sha256 = NULL,
        retain_until = v_completed_at + interval '90 days',
        updated_at = v_completed_at,completed_at = v_completed_at
    WHERE job.id = p_job.id AND job.state = 'running'
      AND job.lease_owner_id = p_lease_owner_id
      AND job.lease_generation = p_lease_generation;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5696',
            MESSAGE = 'event gap job completion failed';
    END IF;
    RETURN v_completed_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_complete_event_gap_job_v1(
    platform.jobs,uuid,uuid,bigint
) FROM PUBLIC;

