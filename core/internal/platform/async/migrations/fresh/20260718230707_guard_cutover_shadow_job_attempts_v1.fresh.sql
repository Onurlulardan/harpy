-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_job_attempt_v2(
    OLD platform.job_attempts, NEW platform.job_attempts,
    p_current_user name, p_session_user name
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_changed jsonb;
    v_receipt_status text;
    v_receipt_found boolean := false;
    v_effect_valid boolean := false;
    v_owner_stale boolean := false;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_session_user IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='P5620',
            MESSAGE='job attempt invocation is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.scope_kind=NEW.scope_kind AND job.scope_id=NEW.scope_id
      AND job.id=NEW.job_id FOR SHARE;
    IF NOT FOUND OR v_job.state IS DISTINCT FROM 'running' OR
       v_job.attempts IS DISTINCT FROM NEW.attempt OR
       v_job.lease_generation IS DISTINCT FROM NEW.lease_generation OR
       v_job.lease_owner_id IS DISTINCT FROM NEW.lease_owner_id THEN
        RAISE EXCEPTION USING ERRCODE='P5621', MESSAGE='job attempt lease is stale';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.cutover_job_bindings AS binding
                   WHERE binding.job_id=v_job.id) THEN
        PERFORM platform.platform_async_validate_job_attempt_v1(
            OLD,NEW,p_current_user,p_session_user);
        RETURN;
    END IF;
    IF OLD IS NULL THEN
        IF NEW.started_at IS DISTINCT FROM v_job.lease_heartbeat_at OR
           NEW.finished_at IS NOT NULL OR NEW.outcome IS NOT NULL OR
           NEW.error_class IS NOT NULL OR NEW.error_code IS NOT NULL OR
           NEW.redacted_error_sha256 IS NOT NULL OR
           NEW.effect_started_at IS NOT NULL OR NEW.effect_id IS NOT NULL OR
           NEW.effect_request_sha256 IS NOT NULL OR
           NEW.effect_receipt_id IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='P5622',
                MESSAGE='job attempt insert is invalid';
        END IF;
        PERFORM platform.platform_async_assert_job_row_authority_v1(v_job);
        RETURN;
    END IF;
    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key) INTO v_changed
    FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING (key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    IF OLD.finished_at IS NULL AND NEW.finished_at IS NULL
       AND OLD.effect_started_at IS NULL AND NEW.effect_started_at IS NOT NULL
       AND OLD.effect_id IS NULL AND NEW.effect_id IS NOT NULL
       AND OLD.effect_request_sha256 IS NULL
       AND NEW.effect_request_sha256 IS NOT NULL
       AND v_job.effect_class='external' AND v_job.cancel_requested_at IS NULL
       AND v_job.effect_started_at IS NOT NULL
       AND NEW.effect_started_at>=NEW.started_at
       AND NEW.effect_started_at>=v_job.effect_started_at
       AND v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
           'effect_started_at','effect_id','effect_request_sha256']::text[]) THEN
        PERFORM platform.platform_async_assert_job_row_authority_v1(v_job);
        RETURN;
    END IF;
    IF OLD.finished_at IS NOT NULL OR NEW.finished_at IS NULL OR
       NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at OR
       (NEW.effect_started_at IS NOT NULL AND
        (v_job.effect_started_at IS NULL OR
         NEW.effect_started_at<v_job.effect_started_at)) OR
       NOT (v_changed OPERATOR(pg_catalog.<@) pg_catalog.to_jsonb(ARRAY[
           'finished_at','outcome','error_class','error_code',
           'redacted_error_sha256','effect_receipt_id']::text[])) THEN
        RAISE EXCEPTION USING ERRCODE='P5623',
            MESSAGE='job attempt transition is forbidden';
    END IF;
    IF NEW.effect_receipt_id IS NOT NULL THEN
        SELECT receipt.effect_status INTO v_receipt_status
        FROM platform.idempotency_receipts AS receipt
        WHERE receipt.receipt_id=NEW.effect_receipt_id
          AND receipt.record_kind='effect'
          AND receipt.scope_kind=v_job.scope_kind AND receipt.scope_id=v_job.scope_id
          AND receipt.authority_group_id=v_job.authority_group_id
          AND receipt.authority_group_generation=v_job.authority_group_generation
          AND receipt.authority_domain=v_job.authority_domain
          AND receipt.cutover_generation=v_job.cutover_generation
          AND receipt.effect_key=v_job.effect_idempotency_key
          AND receipt.effect_id=NEW.effect_id
          AND receipt.request_sha256=NEW.effect_request_sha256
          AND receipt.completed_at<=NEW.finished_at;
        v_receipt_found := FOUND;
    END IF;
    IF v_job.effect_class='database_transactional' THEN
        v_effect_valid := NEW.effect_started_at IS NULL
            AND NEW.effect_receipt_id IS NULL;
    ELSIF NEW.outcome='succeeded' THEN
        v_effect_valid := NEW.effect_started_at IS NOT NULL
            AND v_receipt_found AND v_receipt_status='succeeded';
    ELSIF NEW.outcome='dead' AND NEW.error_code='JOB_OUTCOME_UNKNOWN' THEN
        v_effect_valid := NEW.effect_started_at IS NOT NULL
            AND NEW.effect_receipt_id IS NULL;
    ELSIF NEW.outcome='lease_expired' THEN
        v_effect_valid := (NEW.effect_started_at IS NULL
            AND NEW.effect_receipt_id IS NULL) OR
            (NEW.effect_started_at IS NOT NULL AND v_receipt_found);
    ELSE
        v_effect_valid := (NEW.effect_started_at IS NULL
            AND NEW.effect_receipt_id IS NULL) OR
            (NEW.effect_started_at IS NOT NULL AND v_receipt_found
             AND v_receipt_status='failed');
    END IF;
    IF NOT v_effect_valid THEN
        RAISE EXCEPTION USING ERRCODE='P5624',
            MESSAGE='job attempt effect receipt is invalid';
    END IF;
    v_owner_stale := p_current_user='forgecrud_owner'
        AND NEW.outcome='lease_expired' AND NEW.error_class='conflict'
        AND NEW.error_code='JOB_AUTHORITY_STALE';
    IF NOT v_owner_stale THEN
        PERFORM platform.platform_async_assert_job_row_authority_v1(v_job);
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_validate_job_attempt_v2(
    platform.job_attempts,platform.job_attempts,name,name) FROM PUBLIC;
CREATE FUNCTION platform.platform_async_guard_job_attempt_v5()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog AS $function$
BEGIN
    IF TG_TABLE_SCHEMA IS DISTINCT FROM 'platform'
       OR TG_TABLE_NAME IS DISTINCT FROM 'job_attempts'
       OR NOT (TG_OP=ANY(ARRAY['INSERT','UPDATE']::text[])) THEN
        RAISE EXCEPTION USING ERRCODE='P5620',
            MESSAGE='job attempt invocation is invalid';
    END IF;
    IF TG_OP='INSERT' THEN
        PERFORM platform.platform_async_validate_job_attempt_v2(
            NULL::platform.job_attempts,NEW,CURRENT_USER,SESSION_USER);
    ELSE
        PERFORM platform.platform_async_validate_job_attempt_v2(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_job_attempt_v5() FROM PUBLIC;
CREATE OR REPLACE TRIGGER job_attempts_transition_guard
BEFORE INSERT OR UPDATE ON platform.job_attempts FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_job_attempt_v5();
ALTER TABLE platform.job_attempts ENABLE ALWAYS TRIGGER job_attempts_transition_guard;
