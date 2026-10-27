-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_job_transition_v2(
    OLD platform.jobs, NEW platform.jobs,
    p_current_user name, p_session_user name
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_changed jsonb;
    v_valid boolean;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM platform.cutover_job_bindings AS binding
                   WHERE binding.job_id=OLD.id) OR
       p_current_user='harpy_owner' THEN
        PERFORM platform.platform_async_validate_job_transition_v1(
            OLD,NEW,p_current_user,p_session_user);
        RETURN;
    END IF;
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_session_user IS DISTINCT FROM 'harpy_v2_runtime' OR
       NEW.updated_at<=OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P5610',
            MESSAGE='cutover shadow job transition invocation is invalid';
    END IF;
    PERFORM platform.platform_async_assert_job_row_authority_v1(OLD);
    PERFORM platform.platform_async_assert_job_row_authority_v1(NEW);
    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key) INTO v_changed
    FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING (key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    v_valid :=
        (OLD.state='queued' AND OLD.parked_at IS NULL AND NEW.state='running'
         AND NEW.attempts=OLD.attempts+1
         AND NEW.lease_generation=OLD.lease_generation+1
         AND NEW.lease_heartbeat_at>OLD.updated_at
         AND NEW.lease_until=NEW.lease_heartbeat_at+interval '60 seconds'
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','attempts','lease_generation','lease_owner_id',
             'lease_heartbeat_at','lease_until','updated_at']::text[])) OR
        (OLD.state='running' AND NEW.state='running'
         AND NEW.lease_heartbeat_at>OLD.lease_heartbeat_at
         AND NEW.lease_until=NEW.lease_heartbeat_at+interval '60 seconds'
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'lease_heartbeat_at','lease_until','updated_at']::text[])) OR
        (OLD.state='running' AND NEW.state='running'
         AND OLD.cancel_requested_at IS NULL AND NEW.cancel_requested_at IS NOT NULL
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'cancel_requested_at','updated_at']::text[])) OR
        (OLD.state='running' AND NEW.state='running'
         AND OLD.effect_class='external' AND OLD.effect_started_at IS NULL
         AND OLD.cancel_requested_at IS NULL AND NEW.effect_started_at IS NOT NULL
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'effect_started_at','updated_at']::text[])) OR
        (OLD.state='queued' AND OLD.parked_at IS NULL AND NEW.state='cancelled'
         AND NEW.completed_at=NEW.updated_at AND NEW.last_error_class='cancelled'
         AND NEW.retain_until>=NEW.completed_at+interval '90 days'
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','last_error_class','last_error_code','last_error_sha256',
             'retain_until','completed_at','updated_at']::text[])) OR
        (OLD.state='running' AND NEW.state='queued'
         AND NEW.last_error_class IN ('transient','conflict')
         AND NEW.run_at>=NEW.updated_at AND OLD.attempts<OLD.max_attempts
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'run_at','state','lease_owner_id','lease_heartbeat_at','lease_until',
             'last_error_class','last_error_code','last_error_sha256','updated_at']::text[])) OR
        (OLD.state='running' AND NEW.state='succeeded'
         AND NEW.completed_at=NEW.updated_at AND NEW.last_error_class IS NULL
         AND NEW.last_error_code IS NULL AND NEW.last_error_sha256 IS NULL
         AND NEW.retain_until>=NEW.completed_at+interval '90 days'
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','lease_owner_id','lease_heartbeat_at','lease_until',
             'last_error_class','last_error_code','last_error_sha256',
             'retain_until','completed_at','updated_at']::text[])) OR
        (OLD.state='running' AND NEW.state='dead' AND NEW.completed_at=NEW.updated_at
         AND (NEW.last_error_class IN ('transient','permanent','conflict') OR
              (NEW.last_error_class IS NULL
               AND NEW.last_error_code='JOB_OUTCOME_UNKNOWN'))
         AND NEW.retain_until>=NEW.completed_at+interval '365 days'
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','lease_owner_id','lease_heartbeat_at','lease_until',
             'last_error_class','last_error_code','last_error_sha256',
             'retain_until','completed_at','updated_at']::text[])) OR
        (OLD.state='running' AND NEW.state='cancelled'
         AND NEW.completed_at=NEW.updated_at AND NEW.last_error_class='cancelled'
         AND NEW.retain_until>=NEW.completed_at+interval '90 days'
         AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','lease_owner_id','lease_heartbeat_at','lease_until',
             'last_error_class','last_error_code','last_error_sha256',
             'retain_until','completed_at','updated_at']::text[]));
    IF NOT v_valid THEN
        RAISE EXCEPTION USING ERRCODE='P5613',
            MESSAGE='cutover shadow job transition is forbidden';
    END IF;
    IF OLD.state='running' AND NEW.state<>'running' AND NOT EXISTS (
        SELECT 1 FROM platform.job_attempts AS attempt
        WHERE attempt.job_id=OLD.id AND attempt.attempt=OLD.attempts
          AND attempt.lease_generation=OLD.lease_generation
          AND attempt.lease_owner_id=OLD.lease_owner_id
          AND attempt.finished_at IS NOT NULL
          AND ((NEW.state='queued' AND attempt.outcome IN ('requeued','lease_expired')
                AND attempt.error_class IS NOT DISTINCT FROM NEW.last_error_class
                AND attempt.error_code IS NOT DISTINCT FROM NEW.last_error_code
                AND attempt.redacted_error_sha256 IS NOT DISTINCT FROM
                    NEW.last_error_sha256)
            OR (NEW.state='succeeded' AND attempt.outcome='succeeded')
            OR (NEW.state='dead' AND attempt.outcome='dead'
                AND attempt.error_class IS NOT DISTINCT FROM NEW.last_error_class
                AND attempt.error_code IS NOT DISTINCT FROM NEW.last_error_code
                AND attempt.redacted_error_sha256 IS NOT DISTINCT FROM
                    NEW.last_error_sha256)
            OR (NEW.state='cancelled' AND
                (attempt.outcome='lease_expired' OR
                 (attempt.outcome='cancelled'
                  AND attempt.error_code IS NOT DISTINCT FROM NEW.last_error_code
                  AND attempt.redacted_error_sha256 IS NOT DISTINCT FROM
                      NEW.last_error_sha256))))
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P5612',
            MESSAGE='cutover shadow job transition lacks terminal attempt';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_validate_job_transition_v2(
    platform.jobs,platform.jobs,name,name) FROM PUBLIC;
CREATE FUNCTION platform.platform_async_guard_job_transition_v5()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR TG_TABLE_SCHEMA IS DISTINCT FROM 'platform'
       OR TG_TABLE_NAME IS DISTINCT FROM 'jobs' THEN
        RAISE EXCEPTION USING ERRCODE='P5610',
            MESSAGE='job transition invocation is invalid';
    END IF;
    PERFORM platform.platform_async_validate_job_transition_v2(
        OLD,NEW,CURRENT_USER,SESSION_USER);
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_job_transition_v5()
FROM PUBLIC;
CREATE OR REPLACE TRIGGER jobs_transition_guard
BEFORE UPDATE ON platform.jobs FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_job_transition_v5();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard;
