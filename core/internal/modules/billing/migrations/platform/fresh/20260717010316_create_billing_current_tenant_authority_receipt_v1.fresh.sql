-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_resolve_current_tenant_authority_receipt_v1(
    p_tenant_id uuid
)
RETURNS TABLE(
    authority_id uuid,
    authority_group_id text,
    authority_group_generation bigint,
    billing_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_authority_id uuid;
BEGIN
    v_authority_id :=
        platform.billing_resolve_current_tenant_authority_v1(p_tenant_id);

    RETURN QUERY
    SELECT authority.authority_id,
           authority.authority_group_id,
           authority.authority_group_generation,
           authority.billing_generation
    FROM platform.billing_tenant_authorities AS authority
    WHERE authority.authority_id = v_authority_id
      AND authority.tenant_id = p_tenant_id
      AND authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.authority_group_id = 'billing'
      AND authority.billing_domain = 'billing'
    FOR SHARE OF authority;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing tenant authority receipt is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.billing_resolve_current_tenant_authority_receipt_v1(uuid)
FROM PUBLIC;
