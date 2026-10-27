-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_job_attempt_v3()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_authority_current boolean;
BEGIN
    IF TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'job_attempts' OR
       NOT (TG_OP = ANY (ARRAY['INSERT','UPDATE']::text[])) THEN
        RAISE EXCEPTION USING ERRCODE='P5620',
            MESSAGE='job attempt invocation is invalid';
    END IF;
    IF TG_OP='INSERT' THEN
        PERFORM platform.platform_async_validate_job_attempt_v1(
            NULL::platform.job_attempts,NEW,CURRENT_USER,SESSION_USER);
    ELSIF CURRENT_USER='forgecrud_owner' AND OLD.finished_at IS NULL AND
          NEW.finished_at IS NOT NULL AND OLD.effect_started_at IS NOT NULL AND
          NEW.outcome='dead' AND NEW.error_code='JOB_OUTCOME_UNKNOWN' THEN
        SELECT job.* INTO STRICT v_job FROM platform.jobs AS job
        WHERE job.id=NEW.job_id;
        v_authority_current := EXISTS (
            SELECT 1 FROM platform.cutover_groups AS authority
            WHERE authority.scope_kind=v_job.scope_kind
              AND authority.scope_id=v_job.scope_id
              AND authority.group_id=v_job.authority_group_id
              AND authority.generation=v_job.authority_group_generation
              AND authority.state='v2' AND authority.writer='v2'
        ) AND EXISTS (
            SELECT 1 FROM platform.domain_cutovers AS authority
            WHERE authority.scope_kind=v_job.scope_kind
              AND authority.scope_id=v_job.scope_id
              AND authority.domain=v_job.authority_domain
              AND authority.generation=v_job.cutover_generation
              AND authority.state='v2' AND authority.writer='v2');
        IF v_authority_current THEN
            PERFORM platform.platform_async_validate_job_attempt_v1(
                OLD,NEW,CURRENT_USER,SESSION_USER);
        ELSE
            PERFORM platform.platform_async_validate_stale_external_attempt_unknown_v1(
                OLD,NEW,CURRENT_USER,SESSION_USER);
        END IF;
    ELSE
        PERFORM platform.platform_async_validate_job_attempt_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_guard_job_attempt_v3() FROM PUBLIC;

CREATE OR REPLACE TRIGGER job_attempts_transition_guard
BEFORE INSERT OR UPDATE ON platform.job_attempts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_job_attempt_v3();
ALTER TABLE platform.job_attempts ENABLE ALWAYS TRIGGER job_attempts_transition_guard;

CREATE FUNCTION platform.platform_async_guard_job_transition_v3()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
DECLARE
    v_authority_current boolean;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'jobs' THEN
        RAISE EXCEPTION USING ERRCODE='P5610',
            MESSAGE='job transition invocation is invalid';
    END IF;
    IF CURRENT_USER='forgecrud_owner' AND OLD.parked_at IS NULL AND
       NEW.parked_at IS NOT NULL THEN
        NULL;
    ELSIF CURRENT_USER='forgecrud_owner' AND OLD.state='running' AND
          OLD.effect_started_at IS NOT NULL AND NEW.state='dead' AND
          NEW.last_error_code='JOB_OUTCOME_UNKNOWN' THEN
        v_authority_current := EXISTS (
            SELECT 1 FROM platform.cutover_groups AS authority
            WHERE authority.scope_kind=OLD.scope_kind
              AND authority.scope_id=OLD.scope_id
              AND authority.group_id=OLD.authority_group_id
              AND authority.generation=OLD.authority_group_generation
              AND authority.state='v2' AND authority.writer='v2'
        ) AND EXISTS (
            SELECT 1 FROM platform.domain_cutovers AS authority
            WHERE authority.scope_kind=OLD.scope_kind
              AND authority.scope_id=OLD.scope_id
              AND authority.domain=OLD.authority_domain
              AND authority.generation=OLD.cutover_generation
              AND authority.state='v2' AND authority.writer='v2');
        IF v_authority_current THEN
            PERFORM platform.platform_async_validate_job_transition_v1(
                OLD,NEW,CURRENT_USER,SESSION_USER);
        ELSE
            PERFORM platform.platform_async_validate_stale_external_job_unknown_v1(
                OLD,NEW,CURRENT_USER,SESSION_USER);
        END IF;
    ELSE
        PERFORM platform.platform_async_validate_job_transition_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_guard_job_transition_v3() FROM PUBLIC;

CREATE OR REPLACE TRIGGER jobs_transition_guard
BEFORE UPDATE ON platform.jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_job_transition_v3();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard;
-- +goose StatementEnd
