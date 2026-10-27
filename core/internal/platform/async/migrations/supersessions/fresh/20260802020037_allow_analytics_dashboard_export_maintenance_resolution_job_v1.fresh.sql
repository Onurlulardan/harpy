-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_analytics_export_maint_resolve_job_v1(
    OLD platform.jobs,NEW platform.jobs,
    p_current_user name,p_session_user name
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_attempt platform.job_attempts%ROWTYPE;
    v_receipt platform.idempotency_receipts%ROWTYPE;
    v_resolution platform.analytics_dashboard_export_maintenance_unknown_resolutions%ROWTYPE;
    v_changed jsonb;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'forgecrud_owner' OR
       p_session_user IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       OLD.kind NOT IN ('analytics.dashboard.export.delete',
                        'analytics.dashboard.export.expire') OR
       OLD.state IS DISTINCT FROM 'dead' OR OLD.last_error_class IS NOT NULL OR
       OLD.last_error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       OLD.last_error_sha256 IS NOT NULL OR OLD.completed_at IS NULL OR
       NEW.state IS DISTINCT FROM 'succeeded' OR
       NEW.last_error_class IS NOT NULL OR NEW.last_error_code IS NOT NULL OR
       NEW.last_error_sha256 IS NOT NULL OR
       NEW.completed_at IS DISTINCT FROM NEW.updated_at OR
       NEW.updated_at<=OLD.completed_at OR
       NEW.retain_until IS DISTINCT FROM NEW.completed_at+interval '90 days' OR
       NEW.lease_owner_id IS NOT NULL OR NEW.lease_heartbeat_at IS NOT NULL OR
       NEW.lease_until IS NOT NULL OR
       NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance resolution job is invalid';
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=OLD.id AND attempt.attempt=OLD.attempts
      AND attempt.lease_generation=OLD.lease_generation
      AND attempt.outcome='succeeded' AND attempt.error_class IS NULL
      AND attempt.error_code IS NULL AND attempt.redacted_error_sha256 IS NULL
      AND attempt.effect_started_at=OLD.effect_started_at
      AND attempt.effect_receipt_id IS NOT NULL
      AND attempt.finished_at=NEW.completed_at FOR SHARE;
    SELECT receipt.* INTO v_receipt FROM platform.idempotency_receipts AS receipt
    WHERE receipt.receipt_id=v_attempt.effect_receipt_id
      AND receipt.record_kind='effect' AND receipt.tenant_id=OLD.tenant_id
      AND receipt.scope_kind=OLD.scope_kind AND receipt.scope_id=OLD.scope_id
      AND receipt.authority_group_id=OLD.authority_group_id
      AND receipt.authority_group_generation=OLD.authority_group_generation
      AND receipt.authority_domain=OLD.authority_domain
      AND receipt.cutover_generation=OLD.cutover_generation
      AND receipt.request_sha256=v_attempt.effect_request_sha256
      AND receipt.effect_id=v_attempt.effect_id AND receipt.effect_kind=OLD.kind
      AND receipt.effect_revision=1 AND receipt.effect_key=OLD.effect_idempotency_key
      AND receipt.effect_status='succeeded' AND receipt.effect_error_class IS NULL
      AND receipt.causation_id=OLD.id AND receipt.completed_at=NEW.completed_at
    FOR SHARE;
    SELECT resolution.* INTO v_resolution
    FROM platform.analytics_dashboard_export_maintenance_unknown_resolutions AS resolution
    WHERE resolution.job_id=OLD.id AND resolution.attempt_id=v_attempt.attempt_id
      AND resolution.effect_receipt_id=v_attempt.effect_receipt_id
      AND resolution.kind=OLD.kind AND resolution.tenant_id=OLD.tenant_id
      AND resolution.export_id=OLD.payload_ref::uuid
      AND resolution.original_unknown_at=OLD.completed_at
      AND resolution.original_attempt_finished_at=OLD.completed_at
      AND resolution.original_attempt_outcome='dead'
      AND resolution.original_error_class IS NULL
      AND resolution.original_error_code='JOB_OUTCOME_UNKNOWN'
      AND resolution.original_error_sha256 IS NULL
      AND resolution.resolved_at=NEW.completed_at FOR SHARE;
    IF v_attempt.attempt_id IS NULL OR v_receipt.receipt_id IS NULL OR
       v_resolution.resolution_receipt_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance resolution chain differs';
    END IF;
    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key)
    INTO v_changed FROM (
        SELECT old_field.key
        FROM pg_catalog.jsonb_each(pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING(key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    IF v_changed IS DISTINCT FROM pg_catalog.to_jsonb(ARRAY[
        'completed_at','last_error_code','retain_until','state','updated_at'
    ]::text[]) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance resolution changed evidence';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        OLD.scope_kind,OLD.scope_id,OLD.authority_group_id,OLD.authority_domain);
END;
$function$;

CREATE FUNCTION platform.platform_async_guard_job_transition_v8()
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
    ELSIF OLD.state='dead' AND OLD.last_error_code='JOB_OUTCOME_UNKNOWN' AND
          NEW.state='succeeded' AND OLD.kind IN (
            'analytics.dashboard.export.delete',
            'analytics.dashboard.export.expire') THEN
        PERFORM platform.platform_async_validate_analytics_export_maint_resolve_job_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    ELSE
        PERFORM platform.platform_async_validate_job_transition_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_validate_analytics_export_maint_resolve_job_v1(
        platform.jobs,platform.jobs,name,name) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_guard_job_transition_v8() FROM PUBLIC;
CREATE OR REPLACE TRIGGER jobs_transition_guard
BEFORE UPDATE ON platform.jobs
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_job_transition_v8();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard;
