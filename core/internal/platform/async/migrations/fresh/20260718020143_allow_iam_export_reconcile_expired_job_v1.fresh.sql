-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_iam_export_reconcile_expired_job_v1(
    OLD platform.jobs,NEW platform.jobs,p_current_user name,p_session_user name
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_attempt platform.job_attempts%ROWTYPE;
    v_changed jsonb;
    v_current boolean;
    v_authority_exists boolean;
    v_failure text;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'harpy_owner' OR
       p_session_user IS DISTINCT FROM 'harpy_v2_runtime' OR
       OLD.state<>'running' OR OLD.effect_class<>'external' OR
       OLD.kind<>'iam.system-resource.export-reconcile' OR
       OLD.authority_group_id<>'iam_system_resources' OR
       OLD.authority_domain<>'iam' OR OLD.writer_origin<>'v2' OR
       NEW.state<>'dead' OR NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at OR
       NEW.lease_owner_id IS NOT NULL OR NEW.lease_heartbeat_at IS NOT NULL OR
       NEW.lease_until IS NOT NULL OR NEW.last_error_class<>'permanent' OR
       NEW.last_error_code<>'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED' OR
       NEW.last_error_sha256 IS NULL OR NEW.completed_at<>NEW.updated_at OR
       NEW.updated_at<=OLD.updated_at OR NEW.updated_at<OLD.lease_until OR
       NEW.retain_until<NEW.completed_at+interval '366 days' THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM expired reconciliation job is invalid';
    END IF;
    SELECT EXISTS(
        SELECT 1 FROM platform.cutover_groups AS group_row
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind=group_row.scope_kind
         AND domain_row.scope_id=group_row.scope_id
        WHERE group_row.scope_kind=OLD.scope_kind
          AND group_row.scope_id=OLD.scope_id
          AND group_row.group_id=OLD.authority_group_id
          AND group_row.generation=OLD.authority_group_generation
          AND group_row.state='v2' AND group_row.writer='v2'
          AND domain_row.domain=OLD.authority_domain
          AND domain_row.generation=OLD.cutover_generation
          AND domain_row.state='v2' AND domain_row.writer='v2'
    ),EXISTS(
        SELECT 1 FROM platform.cutover_group_members AS member
        WHERE member.scope_kind=OLD.scope_kind AND member.scope_id=OLD.scope_id
          AND member.group_id=OLD.authority_group_id
          AND member.domain=OLD.authority_domain
    ) INTO v_current,v_authority_exists;
    IF v_current THEN
        PERFORM platform.platform_async_validate_job_transition_v1(
            OLD,NEW,p_current_user,p_session_user);
        RETURN;
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=OLD.id AND attempt.attempt=OLD.attempts
      AND attempt.lease_generation=OLD.lease_generation
      AND attempt.lease_owner_id=OLD.lease_owner_id
      AND attempt.finished_at=NEW.updated_at AND attempt.outcome='dead'
      AND attempt.error_class='permanent'
      AND attempt.error_code=NEW.last_error_code
      AND attempt.redacted_error_sha256=NEW.last_error_sha256
      AND attempt.effect_started_at IS NOT NULL
      AND attempt.effect_receipt_id IS NOT NULL FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM expired reconciliation job attempt differs';
    END IF;
    v_failure := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(
            'iam.system-export.reconciliation-failure.v1','UTF8')||
        pg_catalog.decode('00','hex')||pg_catalog.uuid_send(OLD.id)||
        pg_catalog.uuid_send(v_attempt.effect_id)||pg_catalog.int8send(
            pg_catalog.octet_length(pg_catalog.convert_to(NEW.last_error_code,'UTF8')))||
        pg_catalog.convert_to(NEW.last_error_code,'UTF8')||
        pg_catalog.int8send(OLD.lease_generation)),'hex');
    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key) INTO v_changed
    FROM (
        SELECT old_field.key FROM pg_catalog.jsonb_each(
            pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING(key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    IF NOT v_authority_exists OR NEW.last_error_sha256<>v_failure OR
       v_changed IS DISTINCT FROM pg_catalog.to_jsonb(ARRAY[
           'completed_at','last_error_class','last_error_code','last_error_sha256',
           'lease_heartbeat_at','lease_owner_id','lease_until','retain_until',
           'state','updated_at'
       ]::text[]) OR NOT EXISTS(
           SELECT 1 FROM platform.idempotency_receipts AS receipt
           WHERE receipt.receipt_id=v_attempt.effect_receipt_id
             AND receipt.tenant_id=OLD.tenant_id
             AND receipt.scope_kind=OLD.scope_kind AND receipt.scope_id=OLD.scope_id
             AND receipt.authority_group_id=OLD.authority_group_id
             AND receipt.authority_group_generation=OLD.authority_group_generation
             AND receipt.authority_domain=OLD.authority_domain
             AND receipt.cutover_generation=OLD.cutover_generation
             AND receipt.effect_key=OLD.effect_idempotency_key
             AND receipt.effect_id=v_attempt.effect_id
             AND receipt.request_sha256=v_attempt.effect_request_sha256
             AND receipt.effect_status='failed'
             AND receipt.effect_error_class='permanent'
             AND receipt.effect_code=NEW.last_error_code
             AND receipt.completed_at=NEW.updated_at
       ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM expired reconciliation job evidence differs';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_validate_iam_export_reconcile_expired_job_v1(
        platform.jobs,platform.jobs,name,name
    ) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_guard_job_transition_v3()
RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP<>'UPDATE' OR TG_TABLE_SCHEMA<>'platform' OR TG_TABLE_NAME<>'jobs' THEN
        RAISE EXCEPTION USING ERRCODE='P5610',
            MESSAGE='job transition invocation is invalid';
    END IF;
    IF CURRENT_USER='harpy_owner' AND OLD.parked_at IS NULL
       AND NEW.parked_at IS NOT NULL THEN
        NULL;
    ELSIF CURRENT_USER='harpy_owner' AND OLD.state='running'
       AND NEW.state='dead'
       AND NEW.last_error_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED' THEN
        PERFORM platform.platform_async_validate_iam_export_reconcile_expired_job_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    ELSIF CURRENT_USER='harpy_owner' AND OLD.state='running'
       AND OLD.effect_started_at IS NOT NULL AND NEW.state='dead'
       AND NEW.last_error_code='JOB_OUTCOME_UNKNOWN' THEN
        PERFORM platform.platform_async_validate_stale_external_job_unknown_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    ELSE
        PERFORM platform.platform_async_validate_job_transition_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_guard_job_transition_v3() FROM PUBLIC;
CREATE OR REPLACE TRIGGER jobs_transition_guard BEFORE UPDATE ON platform.jobs
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_job_transition_v3();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard;
-- +goose StatementEnd
