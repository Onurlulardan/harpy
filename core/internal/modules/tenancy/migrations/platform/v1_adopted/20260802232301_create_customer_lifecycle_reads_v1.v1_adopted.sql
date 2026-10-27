-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_list_v1(
    p_environment_id uuid,
    p_limit integer,
    p_after uuid
)
RETURNS TABLE (document jsonb,next_cursor uuid)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_environment_id IS NULL OR p_limit NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Customer tenant lifecycle list denied';
    END IF;
    PERFORM 1 FROM platform.tenancy_deployment_policy AS policy
    WHERE policy.singleton AND policy.environment_id=p_environment_id AND
          policy.deployment_mode='saas';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='SaaS deployment policy is required';
    END IF;
    RETURN QUERY
    WITH candidate AS (
        SELECT tenant.id,tenant.name::text,tenant.slug::text,tenant.status::text,
               tenant.lifecycle_revision,tenant.settings_revision,
               tenant.created_at,tenant.updated_at,
               tenant.offboarded_at,tenant.purged_at
        FROM platform.tenancy_tenant_bindings AS binding
        JOIN platform.tenants AS tenant ON tenant.id=binding.tenant_id
        WHERE binding.environment_id=p_environment_id AND
              binding.purpose='saas_customer' AND
              (p_after IS NULL OR tenant.id > p_after)
        ORDER BY tenant.id
        LIMIT p_limit+1
    ), page AS (
        SELECT * FROM candidate ORDER BY id LIMIT p_limit
    )
    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
               'id',page.id,'name',page.name,'slug',page.slug,
               'status',page.status,'revision',page.lifecycle_revision,
               'settings_revision',page.settings_revision,
               'created_at',page.created_at,'updated_at',page.updated_at,
               'offboarded_at',page.offboarded_at,'purged_at',page.purged_at
           ) ORDER BY page.id),'[]'::jsonb),
           CASE WHEN (SELECT pg_catalog.count(*) FROM candidate)>p_limit
                THEN (SELECT cursor_page.id FROM page AS cursor_page
                      ORDER BY cursor_page.id DESC LIMIT 1)
                ELSE NULL::uuid END
    FROM page;
END;
$function$;

CREATE FUNCTION platform.tenancy_customer_lifecycle_get_v1(
    p_environment_id uuid,
    p_target_tenant_id uuid
)
RETURNS TABLE (
    id uuid,name text,slug text,status text,lifecycle_revision bigint,
    settings_revision bigint,created_at timestamptz,updated_at timestamptz,
    offboarded_at timestamptz,purged_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_environment_id IS NULL OR p_target_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Customer tenant lifecycle read denied';
    END IF;
    RETURN QUERY
    SELECT tenant.id,tenant.name::text,tenant.slug::text,tenant.status::text,
           tenant.lifecycle_revision,tenant.settings_revision,
           tenant.created_at,tenant.updated_at,
           tenant.offboarded_at,tenant.purged_at
    FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id=policy.environment_id AND
         binding.purpose='saas_customer' AND
         binding.tenant_id=p_target_tenant_id
    JOIN platform.tenants AS tenant ON tenant.id=binding.tenant_id
    WHERE policy.singleton AND policy.environment_id=p_environment_id AND
          policy.deployment_mode='saas';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_list_v1(uuid,integer,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_get_v1(uuid,uuid) FROM PUBLIC;
