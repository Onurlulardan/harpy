-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_register_global_authority_v1(
    p_authority_id uuid,
    p_group_generation bigint,
    p_billing_generation bigint,
    p_created_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    environment_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing global registration session is invalid';
    END IF;
    SELECT identity.environment_id INTO environment_id
    FROM platform.environment_identity AS identity
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'global'
     AND group_row.scope_id = identity.environment_id
     AND group_row.group_id = 'billing'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = group_row.scope_kind
     AND domain_row.scope_id = group_row.scope_id
     AND domain_row.domain = 'billing'
    WHERE group_row.state = 'v2' AND group_row.writer = 'v2'
      AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
      AND group_row.generation = p_group_generation
      AND domain_row.generation = p_billing_generation
    FOR SHARE OF group_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing global registration authority is not ready';
    END IF;
    INSERT INTO platform.billing_global_authorities (
        authority_id, scope_kind, scope_id, authority_group_id,
        authority_group_generation, authority_domain,
        authority_domain_generation, created_at
    ) VALUES (
        p_authority_id, 'global', environment_id, 'billing',
        p_group_generation, 'billing', p_billing_generation, p_created_at
    );
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_register_global_authority_v1(
    uuid,bigint,bigint,timestamptz
) FROM PUBLIC;

CREATE FUNCTION platform.billing_register_tenant_authority_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_group_generation bigint,
    p_billing_generation bigint,
    p_forms_generation bigint,
    p_created_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing tenant registration session is invalid';
    END IF;
    PERFORM 1
    FROM platform.cutover_groups AS group_row
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = group_row.scope_kind
     AND billing_row.scope_id = group_row.scope_id
     AND billing_row.domain = 'billing'
    JOIN platform.domain_cutovers AS forms_row
      ON forms_row.scope_kind = group_row.scope_kind
     AND forms_row.scope_id = group_row.scope_id
     AND forms_row.domain = 'forms'
    WHERE group_row.scope_kind = 'tenant'
      AND group_row.scope_id = p_tenant_id
      AND group_row.group_id = 'billing_forms'
      AND group_row.state = 'v2' AND group_row.writer = 'v2'
      AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
      AND forms_row.state = 'v2' AND forms_row.writer = 'v2'
      AND group_row.generation = p_group_generation
      AND billing_row.generation = p_billing_generation
      AND forms_row.generation = p_forms_generation
    FOR SHARE OF group_row, billing_row, forms_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing tenant registration authority is not ready';
    END IF;
    INSERT INTO platform.billing_tenant_authorities (
        authority_id, tenant_id, scope_kind, scope_id, authority_group_id,
        authority_group_generation, billing_domain, billing_generation,
        forms_domain, forms_generation, created_at
    ) VALUES (
        p_authority_id, p_tenant_id, 'tenant', p_tenant_id, 'billing_forms',
        p_group_generation, 'billing', p_billing_generation,
        'forms', p_forms_generation, p_created_at
    );
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,bigint,timestamptz
) FROM PUBLIC;

GRANT USAGE ON SCHEMA platform TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.billing_register_global_authority_v1(
    uuid,bigint,bigint,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.billing_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,bigint,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
