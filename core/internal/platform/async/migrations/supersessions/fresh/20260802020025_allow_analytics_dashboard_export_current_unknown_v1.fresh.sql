-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_analytics_dashboard_export_unknown_attempt_v1(
    OLD platform.job_attempts,NEW platform.job_attempts,
    p_current_user name,p_session_user name
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_changed jsonb;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'harpy_owner' OR
       p_session_user IS DISTINCT FROM 'harpy_v2_runtime' OR
       OLD.finished_at IS NOT NULL OR NEW.finished_at IS NULL OR
       OLD.effect_started_at IS NULL OR
       NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at OR
       NEW.effect_id IS DISTINCT FROM OLD.effect_id OR
       NEW.effect_request_sha256 IS DISTINCT FROM OLD.effect_request_sha256 OR
       NEW.outcome IS DISTINCT FROM 'dead' OR NEW.error_class IS NOT NULL OR
       NEW.error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       NEW.redacted_error_sha256 IS NOT NULL OR
       NEW.effect_receipt_id IS NOT NULL OR
       NEW.finished_at<OLD.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown attempt is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=NEW.job_id AND job.tenant_id=NEW.scope_id
      AND job.scope_kind='tenant' AND job.scope_id=NEW.scope_id
      AND job.authority_domain='analytics' AND job.writer_origin='v2'
      AND job.kind='analytics.dashboard.export.render'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=3 AND job.state='running'
      AND job.attempts=NEW.attempt
      AND job.lease_generation=NEW.lease_generation
      AND job.lease_owner_id=NEW.lease_owner_id
      AND job.effect_started_at=NEW.effect_started_at FOR SHARE;
    IF v_job.id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown attempt job differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    IF EXISTS (
        SELECT 1 FROM platform.idempotency_receipts AS receipt
        WHERE receipt.record_kind='effect'
          AND receipt.scope_kind=v_job.scope_kind
          AND receipt.scope_id=v_job.scope_id
          AND receipt.authority_group_id=v_job.authority_group_id
          AND receipt.authority_group_generation=v_job.authority_group_generation
          AND receipt.authority_domain=v_job.authority_domain
          AND receipt.cutover_generation=v_job.cutover_generation
          AND receipt.effect_key=v_job.effect_idempotency_key
          AND receipt.effect_id=NEW.effect_id
          AND receipt.request_sha256=NEW.effect_request_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown attempt has a receipt';
    END IF;
    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key)
    INTO v_changed FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING(key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    IF v_changed IS DISTINCT FROM pg_catalog.to_jsonb(
        ARRAY['error_code','finished_at','outcome']::text[]
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export unknown changed immutable evidence';
    END IF;
END;
$function$;

CREATE FUNCTION platform.platform_async_guard_job_attempt_v6()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog
AS $function$
DECLARE v_kind text;
BEGIN
    IF TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'job_attempts' OR
       NOT (TG_OP=ANY(ARRAY['INSERT','UPDATE']::text[])) THEN
        RAISE EXCEPTION USING ERRCODE='P5620',
            MESSAGE='job attempt invocation is invalid';
    END IF;
    IF TG_OP='INSERT' THEN
        PERFORM platform.platform_async_validate_job_attempt_v1(
            NULL::platform.job_attempts,NEW,CURRENT_USER,SESSION_USER);
    ELSIF CURRENT_USER='harpy_owner' AND OLD.finished_at IS NULL AND
          NEW.finished_at IS NOT NULL AND OLD.effect_started_at IS NOT NULL AND
          NEW.outcome='dead' AND NEW.error_code='JOB_OUTCOME_UNKNOWN' THEN
        SELECT job.kind INTO v_kind FROM platform.jobs AS job
        WHERE job.id=NEW.job_id;
        IF v_kind='analytics.dashboard.export.render' THEN
            PERFORM platform.platform_async_validate_analytics_dashboard_export_unknown_attempt_v1(
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
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_validate_analytics_dashboard_export_unknown_attempt_v1(
        platform.job_attempts,platform.job_attempts,name,name) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_guard_job_attempt_v6() FROM PUBLIC;

CREATE OR REPLACE TRIGGER job_attempts_transition_guard
BEFORE INSERT OR UPDATE ON platform.job_attempts
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_job_attempt_v6();
