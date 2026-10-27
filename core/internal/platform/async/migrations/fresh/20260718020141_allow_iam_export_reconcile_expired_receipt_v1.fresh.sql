-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_iam_export_reconcile_expired_receipt_v1(
    NEW platform.idempotency_receipts,p_current_user name,p_session_user name
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'harpy_owner' OR
       p_session_user IS DISTINCT FROM 'harpy_v2_runtime' OR
       NEW.record_kind<>'effect' OR NEW.scope_kind<>'tenant' OR
       NEW.tenant_id IS NULL OR NEW.scope_id<>NEW.tenant_id OR
       NEW.authority_group_id<>'iam_system_resources' OR
       NEW.authority_domain<>'iam' OR NEW.writer_origin<>'v2' OR
       NEW.retention_class<>'external_effect' OR
       NEW.effect_kind<>'iam.system-resource.export-reconcile' OR
       NEW.effect_revision<>1 OR NEW.provider<>'object-storage' OR
       NEW.operation<>'head-object' OR NEW.effect_status<>'failed' OR
       NEW.effect_error_class<>'permanent' OR
       NEW.effect_code<>'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED' OR
       NEW.provider_reference IS NOT NULL OR NEW.effect_result_ref IS NOT NULL OR
       NEW.effect_result_sha256 IS NOT NULL OR
       NEW.retain_until<NEW.completed_at+interval '366 days' THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM expired reconciliation receipt is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=NEW.causation_id AND job.tenant_id=NEW.tenant_id
      AND job.scope_kind=NEW.scope_kind AND job.scope_id=NEW.scope_id
      AND job.authority_group_id=NEW.authority_group_id
      AND job.authority_group_generation=NEW.authority_group_generation
      AND job.authority_domain=NEW.authority_domain
      AND job.cutover_generation=NEW.cutover_generation
      AND job.writer_origin='v2'
      AND job.kind='iam.system-resource.export-reconcile'
      AND job.effect_class='external' AND job.effect_idempotency_key=NEW.effect_key
      AND job.state='running' AND job.lease_until<=NEW.completed_at
      AND job.updated_at<NEW.completed_at FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM expired reconciliation receipt job is stale';
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=v_job.id AND attempt.attempt=v_job.attempts
      AND attempt.lease_generation=v_job.lease_generation
      AND attempt.lease_owner_id=v_job.lease_owner_id
      AND attempt.effect_id=NEW.effect_id
      AND attempt.effect_request_sha256=NEW.request_sha256
      AND attempt.effect_started_at IS NOT NULL
      AND attempt.finished_at IS NULL AND attempt.effect_receipt_id IS NULL
    FOR SHARE;
    IF NOT FOUND OR NEW.correlation_id<>NEW.effect_id::text OR
       NEW.completed_at<v_attempt.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM expired reconciliation receipt attempt differs';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_validate_iam_export_reconcile_expired_receipt_v1(
        platform.idempotency_receipts,name,name
    ) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_validate_owner_insert_v6()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP<>'INSERT' OR TG_TABLE_SCHEMA<>'platform' OR NOT (
        TG_TABLE_NAME=ANY(ARRAY[
            'jobs','outbox','idempotency_receipts','dead_letters'
        ]::text[])) THEN
        RAISE EXCEPTION USING ERRCODE='P5604',
            MESSAGE='async owner insert trigger target is invalid';
    END IF;
    IF TG_TABLE_NAME='jobs' THEN
        IF SESSION_USER='harpy_adopter' AND NEW.writer_origin='v1' THEN
            PERFORM platform.platform_async_assert_v1_adopted_job_v1(NEW);
            RETURN NEW;
        END IF;
        IF SESSION_USER='harpy_migrator' AND NEW.writer_origin='v2' THEN
            IF NEW.kind='forms.publication.apply-result' THEN
                PERFORM platform.platform_async_assert_forms_schema_completion_job_v1(NEW);
            ELSIF NEW.kind='forms.schema-evolution.runtime' THEN
                PERFORM platform.platform_async_assert_forms_schema_runtime_job_v2(NEW);
            ELSE
                RAISE EXCEPTION USING ERRCODE='P5605',
                    MESSAGE='migrator async job kind is forbidden';
            END IF;
            RETURN NEW;
        END IF;
        IF SESSION_USER='harpy_v2_runtime' AND NEW.writer_origin='v2'
           AND NEW.kind='forms.schema-evolution.runtime' THEN
            PERFORM platform.platform_async_assert_forms_schema_runtime_job_v2(NEW);
            PERFORM platform.platform_async_assert_current_authority_v1(
                NEW.scope_kind,NEW.scope_id,NEW.authority_group_id,
                NEW.authority_group_generation,NEW.authority_domain,
                NEW.cutover_generation);
            RETURN NEW;
        END IF;
    END IF;
    IF TG_TABLE_NAME='idempotency_receipts' THEN
        IF NEW.effect_kind='iam.system-resource.export-reconcile'
           AND NEW.effect_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED' THEN
            PERFORM platform.platform_async_validate_iam_export_reconcile_expired_receipt_v1(
                NEW,CURRENT_USER,SESSION_USER);
            RETURN NEW;
        END IF;
    END IF;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR NEW.writer_origin IS DISTINCT FROM 'v2' THEN
        RAISE EXCEPTION USING ERRCODE='P5605',
            MESSAGE='async owner insert authority is invalid';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        NEW.scope_kind,NEW.scope_id,NEW.authority_group_id,
        NEW.authority_group_generation,NEW.authority_domain,
        NEW.cutover_generation);
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_validate_owner_insert_v6() FROM PUBLIC;
CREATE OR REPLACE TRIGGER jobs_authority_insert_guard BEFORE INSERT ON platform.jobs
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v6();
CREATE OR REPLACE TRIGGER outbox_authority_insert_guard BEFORE INSERT ON platform.outbox
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v6();
CREATE OR REPLACE TRIGGER idempotency_receipts_authority_insert_guard
BEFORE INSERT ON platform.idempotency_receipts FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v6();
CREATE OR REPLACE TRIGGER dead_letters_authority_insert_guard
BEFORE INSERT ON platform.dead_letters FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v6();
-- +goose StatementEnd
