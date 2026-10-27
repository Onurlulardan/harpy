-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_get_v1(p_operation_id uuid)
RETURNS TABLE (
    job_id uuid,
    operation_id uuid,
    state text,
    tenant_id uuid,
    schema_name text,
    step_plan_sha256 text,
    release_id uuid,
    runtime_generation bigint,
    database_generation bigint,
    migration_manifest_sha256 text,
    error_code text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT job.job_id, job.operation_id, job.state,
           operation.tenant_id, operation.schema_name::text,
           operation.step_plan_sha256, job.release_id,
           job.runtime_generation, job.database_generation,
           job.migration_manifest_sha256, job.error_code
    FROM platform.tenancy_tenant_migration_jobs AS job
    JOIN platform.tenancy_onboarding_operations AS operation
      ON operation.operation_id = job.operation_id
    WHERE job.operation_id = p_operation_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_get_v1(uuid)
FROM PUBLIC;
