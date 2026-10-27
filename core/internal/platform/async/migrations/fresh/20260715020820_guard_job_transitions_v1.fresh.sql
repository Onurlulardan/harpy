-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_job_transition_v1(
    OLD platform.jobs,
    NEW platform.jobs,
    p_current_user name,
    p_session_user name
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_changed jsonb;
    v_normal boolean;
    v_owner boolean;
BEGIN
    IF pg_catalog.pg_trigger_depth() <> 1 OR
       p_session_user IS DISTINCT FROM 'harpy_v2_runtime' OR
       NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P5610',
            MESSAGE = 'job transition invocation is invalid';
    END IF;
    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key)
    INTO v_changed
    FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING (key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    v_normal :=
        (OLD.state = 'queued' AND OLD.parked_at IS NULL AND
         NEW.state = 'running' AND NEW.attempts = OLD.attempts + 1 AND
         NEW.lease_generation = OLD.lease_generation + 1 AND
         NEW.lease_heartbeat_at > OLD.updated_at AND
         NEW.lease_until = NEW.lease_heartbeat_at + interval '60 seconds' AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','attempts','lease_generation','lease_owner_id',
             'lease_heartbeat_at','lease_until','updated_at'
         ]::text[]))) OR
        (OLD.state = 'running' AND NEW.state = 'running' AND
         NEW.lease_heartbeat_at > OLD.lease_heartbeat_at AND
         NEW.lease_until = NEW.lease_heartbeat_at + interval '60 seconds' AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'lease_heartbeat_at','lease_until','updated_at'
         ]::text[]))) OR
        (OLD.state = 'running' AND NEW.state = 'running' AND
         OLD.cancel_requested_at IS NULL AND NEW.cancel_requested_at IS NOT NULL AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'cancel_requested_at','updated_at'
         ]::text[]))) OR
        (OLD.state = 'running' AND NEW.state = 'running' AND
         OLD.effect_class = 'external' AND OLD.effect_started_at IS NULL AND
         OLD.cancel_requested_at IS NULL AND NEW.effect_started_at IS NOT NULL AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'effect_started_at','updated_at'
         ]::text[]))) OR
        (OLD.state = 'queued' AND OLD.parked_at IS NULL AND
         NEW.state = 'cancelled' AND NEW.completed_at = NEW.updated_at AND
         NEW.last_error_class = 'cancelled' AND
         NEW.retain_until >= NEW.completed_at + interval '90 days' AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','last_error_class','last_error_code','last_error_sha256',
             'retain_until','completed_at','updated_at'
         ]::text[]))) OR
        (OLD.state = 'running' AND NEW.state = 'queued' AND
         NEW.last_error_class IN ('transient','conflict') AND
         NEW.run_at >= NEW.updated_at AND OLD.attempts < OLD.max_attempts AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'run_at','state','lease_owner_id','lease_heartbeat_at','lease_until',
             'last_error_class','last_error_code','last_error_sha256','updated_at'
         ]::text[]))) OR
        (OLD.state = 'running' AND NEW.state = 'succeeded' AND
         NEW.completed_at = NEW.updated_at AND NEW.last_error_class IS NULL AND
         NEW.last_error_code IS NULL AND NEW.last_error_sha256 IS NULL AND
         NEW.retain_until >= NEW.completed_at + interval '90 days' AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','lease_owner_id','lease_heartbeat_at','lease_until',
             'last_error_class','last_error_code','last_error_sha256',
             'retain_until','completed_at','updated_at'
         ]::text[]))) OR
        (OLD.state = 'running' AND NEW.state = 'dead' AND
         NEW.completed_at = NEW.updated_at AND
         (NEW.last_error_class IN ('transient','permanent','conflict') OR
          (NEW.last_error_class IS NULL AND
           NEW.last_error_code = 'JOB_OUTCOME_UNKNOWN')) AND
         NEW.retain_until >= NEW.completed_at + interval '365 days' AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','lease_owner_id','lease_heartbeat_at','lease_until',
             'last_error_class','last_error_code','last_error_sha256',
             'retain_until','completed_at','updated_at'
         ]::text[]))) OR
        (OLD.state = 'running' AND NEW.state = 'cancelled' AND
         NEW.completed_at = NEW.updated_at AND NEW.last_error_class = 'cancelled' AND
         NEW.retain_until >= NEW.completed_at + interval '90 days' AND
         (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
             'state','lease_owner_id','lease_heartbeat_at','lease_until',
             'last_error_class','last_error_code','last_error_sha256',
             'retain_until','completed_at','updated_at'
         ]::text[])));
    IF v_normal THEN
        IF platform.platform_async_assert_current_authority_v1(
            OLD.scope_kind, OLD.scope_id, OLD.authority_group_id,
            OLD.authority_group_generation, OLD.authority_domain,
            OLD.cutover_generation
        ) IS NULL THEN NULL; END IF;
        IF OLD.state = 'running' AND NEW.state <> 'running' AND NOT EXISTS (
            SELECT 1 FROM platform.job_attempts AS attempt
            WHERE attempt.job_id = OLD.id AND attempt.attempt = OLD.attempts
              AND attempt.lease_generation = OLD.lease_generation
              AND attempt.lease_owner_id = OLD.lease_owner_id
              AND attempt.finished_at IS NOT NULL
              AND ((NEW.state = 'queued' AND
                    attempt.outcome IN ('requeued','lease_expired') AND
                    attempt.error_class IS NOT DISTINCT FROM NEW.last_error_class AND
                    attempt.error_code IS NOT DISTINCT FROM NEW.last_error_code AND
                    attempt.redacted_error_sha256 IS NOT DISTINCT FROM NEW.last_error_sha256) OR
                   (NEW.state = 'succeeded' AND attempt.outcome = 'succeeded') OR
                   (NEW.state = 'dead' AND attempt.outcome = 'dead' AND
                    attempt.error_class IS NOT DISTINCT FROM NEW.last_error_class AND
                    attempt.error_code IS NOT DISTINCT FROM NEW.last_error_code AND
                    attempt.redacted_error_sha256 IS NOT DISTINCT FROM NEW.last_error_sha256) OR
                   (NEW.state = 'cancelled' AND
                    (attempt.outcome = 'lease_expired' OR
                     (attempt.outcome = 'cancelled' AND
                      attempt.error_code IS NOT DISTINCT FROM NEW.last_error_code AND
                      attempt.redacted_error_sha256 IS NOT DISTINCT FROM NEW.last_error_sha256))))
        ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5612',
                MESSAGE = 'job transition lacks its terminal attempt';
        END IF;
        RETURN;
    END IF;
    v_owner := p_current_user = 'harpy_owner';
    IF v_owner AND OLD.state = 'running' AND NEW.state = 'queued' AND
       OLD.effect_started_at IS NULL AND NEW.parked_at = NEW.updated_at AND
       NEW.parked_code = 'JOB_AUTHORITY_STALE' AND
       NEW.parked_current_group_generation >= OLD.authority_group_generation AND
       NEW.parked_current_cutover_generation >= OLD.cutover_generation AND
       (NEW.parked_current_group_generation > OLD.authority_group_generation OR
        NEW.parked_current_cutover_generation > OLD.cutover_generation) AND
       (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
           'state','lease_owner_id','lease_heartbeat_at','lease_until','parked_at',
           'parked_code','parked_current_group_generation',
           'parked_current_cutover_generation','updated_at'
       ]::text[])) AND EXISTS (
           SELECT 1 FROM platform.job_attempts AS attempt
           WHERE attempt.job_id = OLD.id AND attempt.attempt = OLD.attempts
             AND attempt.lease_generation = OLD.lease_generation
             AND attempt.lease_owner_id = OLD.lease_owner_id
             AND attempt.outcome = 'lease_expired' AND attempt.finished_at IS NOT NULL
       ) THEN
        RETURN;
    END IF;
    IF v_owner AND OLD.state = 'queued' AND OLD.parked_at IS NOT NULL AND
       NEW.state = 'cancelled' AND NEW.completed_at = NEW.updated_at AND
       NEW.last_error_class = 'cancelled' AND
       NEW.retain_until >= NEW.completed_at + interval '90 days' AND
       (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
           'state','last_error_class','last_error_code','last_error_sha256',
           'retain_until','completed_at','updated_at'
       ]::text[])) THEN
        RETURN;
    END IF;
    IF v_owner AND OLD.state IN ('succeeded','dead','cancelled') AND
       NEW.state = OLD.state AND NEW.retention_release_revision =
           OLD.retention_release_revision + 1 AND
       NEW.retention_release_sha256 IS NOT NULL AND
       NEW.retention_released_at = NEW.updated_at AND
       NEW.retain_until >= OLD.retain_until AND
       (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
           'retain_until','retention_release_revision','retention_release_sha256',
           'retention_released_at','updated_at'
       ]::text[])) THEN
        RETURN;
    END IF;
    IF v_owner AND OLD.state IN ('succeeded','dead','cancelled') AND
       OLD.payload_ref IS NOT NULL AND OLD.payload_redacted_at IS NULL AND
       NEW.payload_ref IS NULL AND NEW.payload_redacted_at = NEW.updated_at AND
       NEW.updated_at >= OLD.completed_at + interval '30 days' AND
       (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
           'payload_ref','payload_redacted_at','updated_at'
       ]::text[])) THEN
        RETURN;
    END IF;
    RAISE EXCEPTION USING ERRCODE = 'P5613', MESSAGE = 'job transition is forbidden';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_validate_job_transition_v1(
    platform.jobs,platform.jobs,name,name
) FROM PUBLIC;
