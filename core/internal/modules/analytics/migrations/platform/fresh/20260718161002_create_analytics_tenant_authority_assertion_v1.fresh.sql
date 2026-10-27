-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_assert_tenant_authority_v1(
    p_authority_id uuid,
    p_tenant_id uuid
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_mode text;
    v_state text;
    v_writer text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       pg_catalog.uuid_extract_version(p_authority_id) IS DISTINCT FROM 7 OR
       p_tenant_id IS NULL OR p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics authority assertion input is invalid';
    END IF;
    IF SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime' THEN
        v_mode := 'runtime_v2'; v_state := 'v2'; v_writer := 'v2';
    ELSIF SESSION_USER IS NOT DISTINCT FROM 'forgecrud_adopter' THEN
        v_mode := 'adopter_shadow'; v_state := 'shadow'; v_writer := 'v1';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics authority assertion session is invalid';
    END IF;
    PERFORM 1
    FROM platform.analytics_tenant_authorities AS authority_row
    JOIN platform.tenants AS tenant ON tenant.id = authority_row.tenant_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority_row.scope_kind
     AND group_row.scope_id = authority_row.scope_id
     AND group_row.group_id = authority_row.authority_group_id
    JOIN platform.cutover_group_members AS member_row
      ON member_row.scope_kind = authority_row.scope_kind
     AND member_row.scope_id = authority_row.scope_id
     AND member_row.group_id = authority_row.authority_group_id
     AND member_row.domain = authority_row.authority_domain
     AND member_row.ordinal = 1
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = authority_row.scope_kind
     AND domain_row.scope_id = authority_row.scope_id
     AND domain_row.domain = authority_row.authority_domain
    WHERE authority_row.authority_id = p_authority_id
      AND authority_row.tenant_id = p_tenant_id
      AND tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND authority_row.scope_kind = 'tenant' AND authority_row.scope_id = p_tenant_id
      AND authority_row.authority_group_id = 'analytics'
      AND authority_row.authority_domain = 'analytics'
      AND authority_row.authority_mode = v_mode
      AND authority_row.authority_group_generation = group_row.generation
      AND authority_row.authority_domain_generation = domain_row.generation
      AND group_row.state = v_state AND group_row.writer = v_writer
      AND domain_row.state = v_state AND domain_row.writer = v_writer
    FOR SHARE OF authority_row, tenant, group_row, member_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6101',
            MESSAGE = 'analytics tenant authority is not current';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.analytics_assert_tenant_authority_v1(uuid,uuid) FROM PUBLIC;
-- +goose StatementEnd

