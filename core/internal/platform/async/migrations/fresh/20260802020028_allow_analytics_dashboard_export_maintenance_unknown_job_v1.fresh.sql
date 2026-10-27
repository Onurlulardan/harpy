-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_analytics_dashboard_export_maintenance_unknown_job_v1(
    OLD platform.jobs,NEW platform.jobs,
    p_current_user name,p_session_user name
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_attempt platform.job_attempts%ROWTYPE;
    v_changed jsonb;
    v_max integer;
BEGIN
    v_max:=CASE OLD.kind
        WHEN 'analytics.dashboard.export.delete' THEN 3
        WHEN 'analytics.dashboard.export.expire' THEN 8 END;
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_session_user IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_current_user NOT IN ('forgecrud_owner','forgecrud_runtime',
           'forgecrud_v2_runtime') OR v_max IS NULL OR
       OLD.state IS DISTINCT FROM 'running' OR
       OLD.scope_kind IS DISTINCT FROM 'tenant' OR
       OLD.tenant_id IS DISTINCT FROM OLD.scope_id OR
       OLD.authority_domain IS DISTINCT FROM 'analytics' OR
       OLD.writer_origin IS DISTINCT FROM 'v2' OR
       OLD.payload_revision IS DISTINCT FROM 1 OR
       OLD.effect_class IS DISTINCT FROM 'external' OR
       OLD.max_attempts IS DISTINCT FROM v_max OR
       OLD.attempts NOT BETWEEN 1 AND v_max OR
       OLD.effect_started_at IS NULL OR NEW.state IS DISTINCT FROM 'dead' OR
       NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at OR
       NEW.lease_owner_id IS NOT NULL OR NEW.lease_heartbeat_at IS NOT NULL OR
       NEW.lease_until IS NOT NULL OR NEW.last_error_class IS NOT NULL OR
       NEW.last_error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       NEW.last_error_sha256 IS NOT NULL OR
       NEW.completed_at IS DISTINCT FROM NEW.updated_at OR
       NEW.updated_at<=OLD.updated_at OR
       NEW.retain_until IS DISTINCT FROM NEW.completed_at+interval '365 days' OR
       NOT pg_catalog.pg_input_is_valid(OLD.payload_ref,'uuid') OR
       OLD.payload_ref IS DISTINCT FROM (OLD.payload_ref::uuid)::text OR
       pg_catalog.uuid_extract_version(OLD.payload_ref::uuid) IS DISTINCT FROM 7 OR
       OLD.payload_sha256 IS DISTINCT FROM pg_catalog.encode(pg_catalog.sha256(
           pg_catalog.convert_to(OLD.payload_ref,'UTF8')),'hex') OR
       OLD.dedupe_key IS DISTINCT FROM OLD.kind||':'||OLD.payload_ref OR
       OLD.effect_idempotency_key IS DISTINCT FROM OLD.dedupe_key THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance unknown job is invalid';
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=OLD.id AND attempt.attempt=OLD.attempts
      AND attempt.lease_generation=OLD.lease_generation
      AND attempt.lease_owner_id=OLD.lease_owner_id FOR SHARE;
    IF v_attempt.attempt_id IS NULL OR
       pg_catalog.uuid_extract_version(v_attempt.attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(v_attempt.lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(v_attempt.effect_id) IS DISTINCT FROM 5 OR
       v_attempt.finished_at IS DISTINCT FROM NEW.updated_at OR
       v_attempt.outcome IS DISTINCT FROM 'dead' OR
       v_attempt.error_class IS NOT NULL OR
       v_attempt.error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       v_attempt.redacted_error_sha256 IS NOT NULL OR
       v_attempt.effect_started_at IS DISTINCT FROM OLD.effect_started_at OR
       v_attempt.effect_request_sha256!~'^[0-9a-f]{64}$' OR
       v_attempt.effect_request_sha256=pg_catalog.repeat('0',64) OR
       v_attempt.effect_receipt_id IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance unknown attempt differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        OLD.scope_kind,OLD.scope_id,OLD.authority_group_id,OLD.authority_domain);
    IF EXISTS (
        SELECT 1 FROM platform.idempotency_receipts AS receipt
        WHERE receipt.record_kind='effect' AND receipt.scope_kind=OLD.scope_kind
          AND receipt.scope_id=OLD.scope_id
          AND receipt.authority_group_id=OLD.authority_group_id
          AND receipt.authority_group_generation=OLD.authority_group_generation
          AND receipt.authority_domain=OLD.authority_domain
          AND receipt.cutover_generation=OLD.cutover_generation
          AND receipt.effect_key=OLD.effect_idempotency_key
          AND receipt.effect_id=v_attempt.effect_id
          AND receipt.request_sha256=v_attempt.effect_request_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance unknown has a receipt';
    END IF;
    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key)
    INTO v_changed FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING(key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    IF v_changed IS DISTINCT FROM pg_catalog.to_jsonb(ARRAY[
        'completed_at','last_error_code','lease_heartbeat_at','lease_owner_id',
        'lease_until','retain_until','state','updated_at'
    ]::text[]) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance unknown changed evidence';
    END IF;
END;
$function$;

CREATE FUNCTION platform.platform_async_guard_job_transition_v4()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog
AS $function$
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
    ELSIF OLD.state='running' AND OLD.effect_started_at IS NOT NULL AND
       NEW.state='dead' AND NEW.last_error_code='JOB_OUTCOME_UNKNOWN' THEN
        IF OLD.kind='analytics.dashboard.export.render' THEN
            PERFORM platform.platform_async_validate_analytics_dashboard_export_unknown_job_v1(
                OLD,NEW,CURRENT_USER,SESSION_USER);
        ELSIF OLD.kind IN ('analytics.dashboard.export.delete',
                           'analytics.dashboard.export.expire') THEN
            PERFORM platform.platform_async_validate_analytics_dashboard_export_maintenance_unknown_job_v1(
                OLD,NEW,CURRENT_USER,SESSION_USER);
        ELSIF CURRENT_USER='forgecrud_owner' THEN
            PERFORM platform.platform_async_validate_stale_external_job_unknown_v1(
                OLD,NEW,CURRENT_USER,SESSION_USER);
        ELSE
            PERFORM platform.platform_async_validate_job_transition_v1(
                OLD,NEW,CURRENT_USER,SESSION_USER);
        END IF;
    ELSE
        PERFORM platform.platform_async_validate_job_transition_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_validate_analytics_dashboard_export_maintenance_unknown_job_v1(
        platform.jobs,platform.jobs,name,name) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_guard_job_transition_v4() FROM PUBLIC;
CREATE OR REPLACE TRIGGER jobs_transition_guard
BEFORE UPDATE ON platform.jobs
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_job_transition_v4();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard;
