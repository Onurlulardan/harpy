-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_schema_inspect_v1(
    p_job_id uuid,
    p_attempt integer,
    p_profile_instance_id uuid,
    p_lease_id uuid
)
RETURNS TABLE (
    tenant_id uuid,
    schema_name text,
    namespace_exists boolean,
    schema_receipt_exists boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    RETURN QUERY
    SELECT operation.tenant_id, operation.schema_name::text,
           pg_catalog.to_regnamespace(operation.schema_name::text) IS NOT NULL,
           EXISTS (
               SELECT 1
               FROM platform.tenancy_tenant_migration_receipts AS receipt
               WHERE receipt.job_id = job.job_id
                 AND receipt.receipt_kind = 'schema_created'
           )
    FROM platform.tenancy_tenant_migration_jobs AS job
    JOIN platform.tenancy_onboarding_operations AS operation
      ON operation.operation_id = job.operation_id
    JOIN platform.tenancy_tenant_migration_attempts AS attempt_row
      ON attempt_row.job_id = job.job_id
     AND attempt_row.attempt = job.current_attempt
    WHERE job.job_id = p_job_id AND job.state = 'running'
      AND job.current_attempt = p_attempt
      AND attempt_row.profile_instance_id = p_profile_instance_id
      AND attempt_row.lease_id = p_lease_id
      AND attempt_row.state = 'running'
      AND attempt_row.lease_until > v_now
      AND NOT EXISTS (
          SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = operation.tenant_id
      );
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration schema lease is invalid';
    END IF;
END;
$function$;

CREATE FUNCTION platform.tenancy_tenant_migration_schema_record_v1(
    p_job_id uuid,
    p_attempt integer,
    p_profile_instance_id uuid,
    p_lease_id uuid,
    p_receipt_id uuid,
    p_receipt_sha256 text,
    p_evidence_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_run_id uuid;
    v_schema_name text;
    v_tenant_id uuid;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_receipt_id IS NULL OR p_receipt_sha256 !~ '^[0-9a-f]{64}$' OR
       p_evidence_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration schema receipt is invalid';
    END IF;
    SELECT attempt_row.migration_run_id, operation.schema_name::text,
           operation.tenant_id
    INTO v_run_id, v_schema_name, v_tenant_id
    FROM platform.tenancy_tenant_migration_jobs AS job
    JOIN platform.tenancy_onboarding_operations AS operation
      ON operation.operation_id = job.operation_id
    JOIN platform.tenancy_tenant_migration_attempts AS attempt_row
      ON attempt_row.job_id = job.job_id
     AND attempt_row.attempt = job.current_attempt
    WHERE job.job_id = p_job_id AND job.state = 'running'
      AND job.current_attempt = p_attempt
      AND attempt_row.profile_instance_id = p_profile_instance_id
      AND attempt_row.lease_id = p_lease_id
      AND attempt_row.state = 'running'
      AND attempt_row.lease_until > v_now
    FOR UPDATE OF job, attempt_row;
    IF NOT FOUND OR pg_catalog.to_regnamespace(v_schema_name) IS NULL OR
       EXISTS (SELECT 1 FROM platform.tenants WHERE id = v_tenant_id) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration schema evidence is invalid';
    END IF;
    INSERT INTO platform.tenancy_tenant_migration_receipts (
        receipt_id, job_id, attempt, receipt_kind, migration_run_id,
        receipt_sha256, evidence_sha256, occurred_at
    ) VALUES (
        p_receipt_id, p_job_id, p_attempt, 'schema_created', v_run_id,
        p_receipt_sha256, p_evidence_sha256, v_now
    );
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_schema_inspect_v1(
    uuid, integer, uuid, uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_schema_record_v1(
    uuid, integer, uuid, uuid, uuid, text, text
) FROM PUBLIC;
