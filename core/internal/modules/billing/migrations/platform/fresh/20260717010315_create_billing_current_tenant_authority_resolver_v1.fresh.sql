-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_resolve_current_tenant_authority_v1(
    p_tenant_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_authority_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL OR
       p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing tenant authority resolver input is invalid';
    END IF;
    BEGIN
        SELECT authority.authority_id INTO STRICT v_authority_id
        FROM platform.billing_tenant_authorities AS authority
        JOIN platform.tenants AS tenant ON tenant.id = authority.tenant_id
        JOIN platform.cutover_groups AS group_row
          ON group_row.scope_kind = authority.scope_kind
         AND group_row.scope_id = authority.scope_id
         AND group_row.group_id = authority.authority_group_id
        JOIN platform.domain_cutovers AS billing_row
          ON billing_row.scope_kind = authority.scope_kind
         AND billing_row.scope_id = authority.scope_id
         AND billing_row.domain = authority.billing_domain
        JOIN platform.domain_cutovers AS forms_row
          ON forms_row.scope_kind = authority.scope_kind
         AND forms_row.scope_id = authority.scope_id
         AND forms_row.domain = authority.forms_domain
        JOIN platform.cutover_group_members AS billing_member
          ON billing_member.scope_kind = authority.scope_kind
         AND billing_member.scope_id = authority.scope_id
         AND billing_member.group_id = authority.authority_group_id
         AND billing_member.domain = 'billing' AND billing_member.ordinal = 1
        JOIN platform.cutover_group_members AS forms_member
          ON forms_member.scope_kind = authority.scope_kind
         AND forms_member.scope_id = authority.scope_id
         AND forms_member.group_id = authority.authority_group_id
         AND forms_member.domain = 'forms' AND forms_member.ordinal = 2
        WHERE authority.tenant_id = p_tenant_id
          AND authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
          AND authority.authority_group_id = 'billing_forms'
          AND authority.billing_domain = 'billing' AND authority.forms_domain = 'forms'
          AND tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE', 'TRIAL')
          AND group_row.state = 'v2' AND group_row.writer = 'v2'
          AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
          AND forms_row.state = 'v2' AND forms_row.writer = 'v2'
          AND authority.authority_group_generation = group_row.generation
          AND authority.billing_generation = billing_row.generation
          AND authority.forms_generation = forms_row.generation
          AND NOT EXISTS (
              SELECT 1 FROM platform.cutover_group_members AS extra
              WHERE extra.scope_kind = 'tenant' AND extra.scope_id = p_tenant_id
                AND extra.group_id = 'billing_forms'
                AND NOT ((extra.ordinal = 1 AND extra.domain = 'billing')
                      OR (extra.ordinal = 2 AND extra.domain = 'forms'))
          )
        FOR SHARE OF authority, tenant, group_row, billing_row, forms_row,
                     billing_member, forms_member;
    EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing+forms tenant authority is not current';
    END;
    PERFORM platform.billing_assert_tenant_authority_v1(v_authority_id);
    RETURN v_authority_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.billing_resolve_current_tenant_authority_v1(uuid) FROM PUBLIC;
