-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_schema_authority_current_v2(
    p_tenant_id uuid, p_authority_id uuid, p_authority_group_id text,
    p_group_generation bigint, p_billing_generation bigint,
    p_forms_generation bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_migrator'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.uuid_extract_version(p_tenant_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_authority_id) IS DISTINCT FROM 7
       OR p_authority_group_id IS DISTINCT FROM 'billing_forms'
       OR p_group_generation NOT BETWEEN 1 AND 9007199254740991
       OR p_billing_generation NOT BETWEEN 1 AND 9007199254740991
       OR p_forms_generation NOT BETWEEN 1 AND 9007199254740991 THEN
        RETURN false;
    END IF;
    PERFORM 1
    FROM platform.billing_tenant_authorities authority
    JOIN platform.tenants tenant ON tenant.id=authority.tenant_id
    JOIN platform.cutover_groups group_row
      ON group_row.scope_kind=authority.scope_kind
     AND group_row.scope_id=authority.scope_id
     AND group_row.group_id=authority.authority_group_id
    JOIN platform.domain_cutovers billing_row
      ON billing_row.scope_kind=authority.scope_kind
     AND billing_row.scope_id=authority.scope_id
     AND billing_row.domain=authority.billing_domain
    JOIN platform.domain_cutovers forms_row
      ON forms_row.scope_kind=authority.scope_kind
     AND forms_row.scope_id=authority.scope_id
     AND forms_row.domain=authority.forms_domain
    JOIN platform.cutover_group_members billing_member
      ON billing_member.scope_kind=authority.scope_kind
     AND billing_member.scope_id=authority.scope_id
     AND billing_member.group_id=authority.authority_group_id
     AND billing_member.domain='billing' AND billing_member.ordinal=1
    JOIN platform.cutover_group_members forms_member
      ON forms_member.scope_kind=authority.scope_kind
     AND forms_member.scope_id=authority.scope_id
     AND forms_member.group_id=authority.authority_group_id
     AND forms_member.domain='forms' AND forms_member.ordinal=2
    WHERE authority.authority_id=p_authority_id
      AND authority.tenant_id=p_tenant_id
      AND authority.scope_kind='tenant' AND authority.scope_id=p_tenant_id
      AND authority.authority_group_id=p_authority_group_id
      AND authority.authority_group_generation=p_group_generation
      AND authority.billing_domain='billing'
      AND authority.billing_generation=p_billing_generation
      AND authority.forms_domain='forms'
      AND authority.forms_generation=p_forms_generation
      AND tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE','TRIAL')
      AND group_row.state='v2' AND group_row.writer='v2'
      AND group_row.generation=p_group_generation
      AND billing_row.state='v2' AND billing_row.writer='v2'
      AND billing_row.generation=p_billing_generation
      AND forms_row.state='v2' AND forms_row.writer='v2'
      AND forms_row.generation=p_forms_generation
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members extra
          WHERE extra.scope_kind='tenant' AND extra.scope_id=p_tenant_id
            AND extra.group_id='billing_forms'
            AND NOT ((extra.ordinal=1 AND extra.domain='billing')
                  OR (extra.ordinal=2 AND extra.domain='forms'))
      )
    FOR SHARE OF authority,tenant,group_row,billing_row,forms_row,
                 billing_member,forms_member;
    RETURN FOUND;
EXCEPTION WHEN OTHERS THEN
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_migrator_schema_authority_current_v2(
    uuid,uuid,text,bigint,bigint,bigint
) FROM PUBLIC;
