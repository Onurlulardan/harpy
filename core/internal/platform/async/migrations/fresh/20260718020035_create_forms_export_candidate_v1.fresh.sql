-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_candidate_v1()
RETURNS TABLE(
    tenant_id uuid,
    tenant_schema name,
    export_id uuid,
    job_id uuid,
    attempt_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms export candidate authority is invalid';
    END IF;
    RETURN QUERY
    SELECT job.tenant_id,tenant.schema_name::name,
           pg_catalog.substring(job.payload_ref,21)::uuid,
           job.id,(job.attempts + 1)::bigint
    FROM platform.jobs AS job
    JOIN platform.tenants AS tenant ON tenant.id = job.tenant_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = job.scope_kind
     AND group_row.scope_id = job.scope_id
     AND group_row.group_id = job.authority_group_id
     AND group_row.generation = job.authority_group_generation
     AND group_row.state = 'v2' AND group_row.writer = 'v2'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = job.scope_kind
     AND domain_row.scope_id = job.scope_id
     AND domain_row.domain = job.authority_domain
     AND domain_row.generation = job.cutover_generation
     AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
    WHERE job.scope_kind = 'tenant' AND job.scope_id = job.tenant_id
      AND job.authority_domain = 'forms' AND job.writer_origin = 'v2'
      AND job.kind = 'forms.record-export' AND job.payload_revision = 1
      AND job.effect_class = 'external' AND job.max_attempts = 3
      AND job.payload_ref = 'forms-record-export/' ||
          pg_catalog.substring(job.payload_ref,21)
      AND job.state = 'queued' AND job.parked_at IS NULL
      AND job.attempts < job.max_attempts
      AND job.run_at <= pg_catalog.clock_timestamp()
      AND tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE','TRIAL')
    ORDER BY job.priority DESC,job.run_at,job.id
    LIMIT 1;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_forms_export_candidate_v1() FROM PUBLIC;

CREATE FUNCTION platform.platform_async_forms_export_claimable_tenant_v1(
    p_tenant_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE v_claimable boolean;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Forms export tenant preflight authority is invalid';
    END IF;
    SELECT true INTO v_claimable FROM platform.tenants AS tenant
    WHERE tenant.id=p_tenant_id AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE','TRIAL') FOR SHARE;
    RETURN COALESCE(v_claimable,false);
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_forms_export_claimable_tenant_v1(uuid) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_forms_export_kek_tenant_v1(
    p_after_tenant_id uuid
)
RETURNS TABLE(tenant_id uuid, tenant_schema name)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms export KEK inventory authority is invalid';
    END IF;
    RETURN QUERY
    SELECT tenant.id,tenant.schema_name::name
    FROM platform.tenants AS tenant
    JOIN platform.domain_cutovers AS authority
      ON authority.scope_kind = 'tenant' AND authority.scope_id = tenant.id
     AND authority.domain = 'forms' AND authority.state = 'v2'
     AND authority.writer = 'v2'
    WHERE tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE','TRIAL')
      AND (p_after_tenant_id IS NULL OR tenant.id > p_after_tenant_id)
    ORDER BY tenant.id
    LIMIT 1;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_forms_export_kek_tenant_v1(uuid) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_forms_export_kek_schema_v1(
    p_tenant_id uuid
)
RETURNS name
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE v_schema text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Forms export KEK schema authority is invalid';
    END IF;
    SELECT tenant.schema_name::text INTO v_schema
    FROM platform.tenants AS tenant
    JOIN platform.domain_cutovers AS authority
      ON authority.scope_kind='tenant' AND authority.scope_id=tenant.id
     AND authority.domain='forms' AND authority.state='v2'
     AND authority.writer='v2'
    WHERE tenant.id=p_tenant_id AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE','TRIAL');
    RETURN v_schema::name;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_forms_export_kek_schema_v1(uuid) FROM PUBLIC;
-- +goose StatementEnd
