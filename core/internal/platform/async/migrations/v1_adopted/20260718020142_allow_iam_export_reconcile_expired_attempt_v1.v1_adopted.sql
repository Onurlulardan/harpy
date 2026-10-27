-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_iam_export_reconcile_expired_attempt_v1(
    OLD platform.job_attempts,NEW platform.job_attempts,
    p_current_user name,p_session_user name
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_changed jsonb;
    v_current boolean;
    v_authority_exists boolean;
    v_failure text;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'harpy_owner' OR
       p_session_user IS DISTINCT FROM 'harpy_v2_runtime' OR
       OLD.finished_at IS NOT NULL OR NEW.finished_at IS NULL OR
       NEW.outcome<>'dead' OR NEW.error_class<>'permanent' OR
       NEW.error_code<>'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED' OR
       NEW.redacted_error_sha256 IS NULL OR NEW.effect_receipt_id IS NULL OR
       NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at OR
       NEW.effect_id IS DISTINCT FROM OLD.effect_id OR
       NEW.effect_request_sha256 IS DISTINCT FROM OLD.effect_request_sha256 OR
       NEW.finished_at<OLD.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM expired reconciliation attempt is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=NEW.job_id AND job.scope_kind=NEW.scope_kind
      AND job.scope_id=NEW.scope_id AND job.tenant_id=NEW.scope_id
      AND job.authority_group_id='iam_system_resources'
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.kind='iam.system-resource.export-reconcile'
      AND job.effect_class='external' AND job.state='running'
      AND job.attempts=NEW.attempt AND job.lease_generation=NEW.lease_generation
      AND job.lease_owner_id=NEW.lease_owner_id
      AND job.effect_started_at IS NOT NULL
      AND job.lease_until<=NEW.finished_at AND job.updated_at<NEW.finished_at
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM expired reconciliation attempt lease is stale';
    END IF;
    SELECT EXISTS(
        SELECT 1 FROM platform.cutover_groups AS group_row
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind=group_row.scope_kind
         AND domain_row.scope_id=group_row.scope_id
        WHERE group_row.scope_kind=v_job.scope_kind
          AND group_row.scope_id=v_job.scope_id
          AND group_row.group_id=v_job.authority_group_id
          AND group_row.generation=v_job.authority_group_generation
          AND group_row.state='v2' AND group_row.writer='v2'
          AND domain_row.domain=v_job.authority_domain
          AND domain_row.generation=v_job.cutover_generation
          AND domain_row.state='v2' AND domain_row.writer='v2'
    ),EXISTS(
        SELECT 1 FROM platform.cutover_group_members AS member
        WHERE member.scope_kind=v_job.scope_kind AND member.scope_id=v_job.scope_id
          AND member.group_id=v_job.authority_group_id
          AND member.domain=v_job.authority_domain
    ) INTO v_current,v_authority_exists;
    IF v_current THEN
        PERFORM platform.platform_async_validate_job_attempt_v1(
            OLD,NEW,p_current_user,p_session_user);
        RETURN;
    END IF;
    v_failure := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(
            'iam.system-export.reconciliation-failure.v1','UTF8')||
        pg_catalog.decode('00','hex')||pg_catalog.uuid_send(NEW.job_id)||
        pg_catalog.uuid_send(NEW.effect_id)||pg_catalog.int8send(
            pg_catalog.octet_length(pg_catalog.convert_to(NEW.error_code,'UTF8')))||
        pg_catalog.convert_to(NEW.error_code,'UTF8')||
        pg_catalog.int8send(NEW.lease_generation)),'hex');
    SELECT pg_catalog.jsonb_agg(changed.key ORDER BY changed.key) INTO v_changed
    FROM (
        SELECT old_field.key FROM pg_catalog.jsonb_each(
            pg_catalog.to_jsonb(OLD)) AS old_field
        JOIN pg_catalog.jsonb_each(pg_catalog.to_jsonb(NEW)) AS new_field USING(key)
        WHERE old_field.value IS DISTINCT FROM new_field.value
    ) AS changed;
    IF NOT v_authority_exists OR NEW.redacted_error_sha256<>v_failure OR
       v_changed IS DISTINCT FROM pg_catalog.to_jsonb(ARRAY[
           'effect_receipt_id','error_class','error_code','finished_at',
           'outcome','redacted_error_sha256'
       ]::text[]) OR NOT EXISTS(
           SELECT 1 FROM platform.idempotency_receipts AS receipt
           WHERE receipt.receipt_id=NEW.effect_receipt_id
             AND receipt.scope_kind=NEW.scope_kind AND receipt.scope_id=NEW.scope_id
             AND receipt.authority_group_id=v_job.authority_group_id
             AND receipt.authority_group_generation=v_job.authority_group_generation
             AND receipt.authority_domain=v_job.authority_domain
             AND receipt.cutover_generation=v_job.cutover_generation
             AND receipt.effect_key=v_job.effect_idempotency_key
             AND receipt.effect_id=NEW.effect_id
             AND receipt.request_sha256=NEW.effect_request_sha256
             AND receipt.effect_status='failed'
             AND receipt.effect_error_class='permanent'
             AND receipt.effect_code=NEW.error_code
             AND receipt.completed_at=NEW.finished_at
       ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM expired reconciliation attempt evidence differs';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_validate_iam_export_reconcile_expired_attempt_v1(
        platform.job_attempts,platform.job_attempts,name,name
    ) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_guard_job_attempt_v3()
RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_TABLE_SCHEMA<>'platform' OR TG_TABLE_NAME<>'job_attempts'
       OR NOT (TG_OP=ANY(ARRAY['INSERT','UPDATE']::text[])) THEN
        RAISE EXCEPTION USING ERRCODE='P5620',
            MESSAGE='job attempt invocation is invalid';
    END IF;
    IF TG_OP='INSERT' THEN
        PERFORM platform.platform_async_validate_job_attempt_v1(
            NULL::platform.job_attempts,NEW,CURRENT_USER,SESSION_USER);
    ELSIF CURRENT_USER='harpy_owner' AND OLD.finished_at IS NULL
       AND NEW.finished_at IS NOT NULL AND NEW.outcome='dead'
       AND NEW.error_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED' THEN
        PERFORM platform.platform_async_validate_iam_export_reconcile_expired_attempt_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    ELSIF CURRENT_USER='harpy_owner' AND OLD.finished_at IS NULL
       AND NEW.finished_at IS NOT NULL AND OLD.effect_started_at IS NOT NULL
       AND NEW.outcome='dead' AND NEW.error_code='JOB_OUTCOME_UNKNOWN' THEN
        PERFORM platform.platform_async_validate_stale_external_attempt_unknown_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    ELSE
        PERFORM platform.platform_async_validate_job_attempt_v1(
            OLD,NEW,CURRENT_USER,SESSION_USER);
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_guard_job_attempt_v3() FROM PUBLIC;
CREATE OR REPLACE TRIGGER job_attempts_transition_guard
BEFORE INSERT OR UPDATE ON platform.job_attempts FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_job_attempt_v3();
ALTER TABLE platform.job_attempts ENABLE ALWAYS TRIGGER job_attempts_transition_guard;
-- +goose StatementEnd
