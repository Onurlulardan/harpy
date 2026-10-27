-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_complete_v3(
    p_job_id uuid, p_attempt integer, p_profile_instance_id uuid,
    p_lease_id uuid, p_receipt_id uuid, p_receipt_sha256 text,
    p_evidence_sha256 text
)
RETURNS TABLE (catalog_sha256 text, requirement_count integer)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_catalog_sha256 text;
    v_requirement_count integer;
    v_operation_id uuid;
    v_leased_at timestamptz;
    v_attempt_number bigint;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_tenant_migration_executor'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='tenant migration completion authority is denied';
    END IF;
    SELECT completed.catalog_sha256,completed.requirement_count
    INTO v_catalog_sha256,v_requirement_count
    FROM platform.tenancy_tenant_migration_complete_v2(
        p_job_id,p_attempt,p_profile_instance_id,p_lease_id,p_receipt_id,
        p_receipt_sha256,p_evidence_sha256
    ) AS completed;
    SELECT job.operation_id INTO STRICT v_operation_id
    FROM platform.tenancy_tenant_migration_jobs AS job
    WHERE job.job_id=p_job_id AND job.state='succeeded'
    FOR SHARE;
    SELECT attempt.leased_at INTO STRICT v_leased_at
    FROM platform.tenancy_tenant_migration_attempts AS attempt
    WHERE attempt.job_id=p_job_id AND attempt.attempt=p_attempt
      AND attempt.profile_instance_id=p_profile_instance_id
      AND attempt.lease_id=p_lease_id AND attempt.state='succeeded';
    SELECT COALESCE(MAX(attempt.attempt_number),0)+1 INTO v_attempt_number
    FROM platform.tenancy_onboarding_attempts AS attempt
    WHERE attempt.operation_id=v_operation_id;
    INSERT INTO platform.tenancy_onboarding_attempts (
        attempt_id,operation_id,attempt_number,outcome,started_at,finished_at
    ) VALUES (
        p_lease_id,v_operation_id,v_attempt_number,'completed',v_leased_at,v_now
    );
    INSERT INTO platform.tenancy_onboarding_receipts (
        receipt_id,operation_id,attempt_id,receipt_kind,
        receipt_sha256,evidence_sha256,occurred_at
    ) VALUES (
        p_receipt_id,v_operation_id,p_lease_id,'migration_completed',
        p_receipt_sha256,p_evidence_sha256,v_now
    );
    UPDATE platform.tenancy_onboarding_operations AS operation
    SET state='migration_ready',revision=operation.revision+1,updated_at=v_now
    WHERE operation.operation_id=v_operation_id
      AND operation.operation_kind='saas_customer_tenant_provision'
      AND operation.state='migrating';
    RETURN QUERY SELECT v_catalog_sha256,v_requirement_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_complete_v3(
    uuid,integer,uuid,uuid,uuid,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_tenant_migration_complete_v3(
    uuid,integer,uuid,uuid,uuid,text,text
) TO harpy_tenant_migration_executor;
