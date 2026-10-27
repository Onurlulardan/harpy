-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_register_tenant_authority_v1(
    p_authority_id uuid, p_tenant_id uuid,
    p_group_generation bigint, p_domain_generation bigint,
    p_created_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_mode text;
    v_state text;
    v_writer text;
    v_inserted boolean;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       pg_catalog.uuid_extract_version(p_authority_id) IS DISTINCT FROM 7 OR
       p_tenant_id IS NULL OR p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_group_generation IS NULL OR p_group_generation < 1 OR
       p_domain_generation IS NULL OR p_domain_generation < 1 OR
       p_created_at IS NULL OR NOT pg_catalog.isfinite(p_created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics authority registration input is invalid';
    END IF;
    IF SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime' THEN
        v_mode := 'runtime_v2'; v_state := 'v2'; v_writer := 'v2';
    ELSIF SESSION_USER IS NOT DISTINCT FROM 'harpy_adopter' THEN
        v_mode := 'adopter_shadow'; v_state := 'shadow'; v_writer := 'v1';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics authority registration session is invalid';
    END IF;
    PERFORM 1
    FROM platform.tenants AS tenant
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'tenant' AND group_row.scope_id = tenant.id
     AND group_row.group_id = 'analytics'
    JOIN platform.cutover_group_members AS member_row
      ON member_row.scope_kind = group_row.scope_kind
     AND member_row.scope_id = group_row.scope_id
     AND member_row.group_id = group_row.group_id
     AND member_row.ordinal = 1 AND member_row.domain = 'analytics'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = member_row.scope_kind
     AND domain_row.scope_id = member_row.scope_id
     AND domain_row.domain = member_row.domain
    WHERE tenant.id = p_tenant_id AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND group_row.generation = p_group_generation
      AND domain_row.generation = p_domain_generation
      AND group_row.state = v_state AND group_row.writer = v_writer
      AND domain_row.state = v_state AND domain_row.writer = v_writer
    FOR SHARE OF tenant, group_row, member_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6101',
            MESSAGE = 'analytics authority registration is not current';
    END IF;
    INSERT INTO platform.analytics_tenant_authorities (
        authority_id, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, authority_domain_generation, authority_mode, created_at
    ) VALUES (
        p_authority_id, p_tenant_id, 'tenant', p_tenant_id,
        'analytics', p_group_generation, 'analytics', p_domain_generation,
        v_mode, p_created_at
    ) ON CONFLICT DO NOTHING RETURNING true INTO v_inserted;
    IF NOT COALESCE(v_inserted, false) AND NOT EXISTS (
        SELECT 1 FROM platform.analytics_tenant_authorities AS authority_row
        WHERE authority_row.authority_id = p_authority_id
          AND authority_row.tenant_id = p_tenant_id
          AND authority_row.authority_group_generation = p_group_generation
          AND authority_row.authority_domain_generation = p_domain_generation
          AND authority_row.authority_mode = v_mode
          AND authority_row.created_at IS NOT DISTINCT FROM p_created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P6101',
            MESSAGE = 'analytics authority registration replay differs';
    END IF;
    PERFORM platform.analytics_assert_tenant_authority_v1(p_authority_id, p_tenant_id);
    RETURN p_authority_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.analytics_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.analytics_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,timestamptz
) TO harpy_adopter;

