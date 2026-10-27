-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(
    p_after_tenant_id uuid
)
RETURNS TABLE(tenant_id uuid,tenant_schema name)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path=pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM import cleanup inventory authority is invalid';
    END IF;
    RETURN QUERY
    SELECT tenant.id,tenant.schema_name::name
    FROM platform.tenants AS tenant
    JOIN platform.domain_cutovers AS authority
      ON authority.scope_kind='tenant' AND authority.scope_id=tenant.id
     AND authority.domain='iam' AND authority.state='v2'
     AND authority.writer='v2'
    WHERE tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE','TRIAL')
      AND (p_after_tenant_id IS NULL OR tenant.id>p_after_tenant_id)
    ORDER BY tenant.id
    LIMIT 1;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_iam_import_cleanup_tenant_v1(uuid) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_iam_import_cleanup_read_v1(
    p_tenant_id uuid,p_import_id uuid,p_job_id uuid,p_resource_kind text
)
RETURNS TABLE(job_state text,completed_at timestamptz)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path=pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_resource_kind NOT IN ('users','roles','organizations') THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM import cleanup job read authority is invalid';
    END IF;
    RETURN QUERY
    SELECT job.state::text,job.completed_at
    FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.authority_group_id='iam_system_resources'
      AND job.kind='iam.system-resource.record-import'
      AND job.payload_revision=1
      AND job.effect_class='database_transactional'
      AND job.effect_idempotency_key IS NULL
      AND job.dedupe_key='iam-system-transfer:'||
          'iam.system-resource.record-import:'||p_import_id::text
      AND (job.payload_ref='iam-system-record/'||p_resource_kind||'/'||
           p_import_id::text OR
           (job.payload_ref IS NULL AND job.payload_redacted_at IS NOT NULL))
      AND job.max_attempts=8;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import cleanup job was not found';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_iam_import_cleanup_read_v1(
    uuid,uuid,uuid,text
) FROM PUBLIC;
-- +goose StatementEnd
