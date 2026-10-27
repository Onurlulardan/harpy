-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_guard_tenant_authority_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_mode text;
    v_state text;
    v_writer text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics tenant authority mutation is invalid';
    END IF;
    IF SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime' THEN
        v_mode := 'runtime_v2'; v_state := 'v2'; v_writer := 'v2';
    ELSIF SESSION_USER IS NOT DISTINCT FROM 'harpy_adopter' THEN
        v_mode := 'adopter_shadow'; v_state := 'shadow'; v_writer := 'v1';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics tenant authority session is invalid';
    END IF;
    IF NEW.authority_mode IS DISTINCT FROM v_mode THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics tenant authority mode is invalid';
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
    WHERE tenant.id = NEW.tenant_id AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND NEW.scope_kind = 'tenant' AND NEW.scope_id = tenant.id
      AND NEW.authority_group_id = group_row.group_id
      AND NEW.authority_domain = domain_row.domain
      AND NEW.authority_group_generation = group_row.generation
      AND NEW.authority_domain_generation = domain_row.generation
      AND group_row.state = v_state AND group_row.writer = v_writer
      AND domain_row.state = v_state AND domain_row.writer = v_writer
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS extra
          WHERE extra.scope_kind = 'tenant' AND extra.scope_id = tenant.id
            AND extra.group_id = 'analytics'
            AND NOT (extra.ordinal = 1 AND extra.domain = 'analytics')
      )
    FOR SHARE OF tenant, group_row, member_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6101',
            MESSAGE = 'analytics tenant authority binding is stale';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.analytics_guard_tenant_authority_v1() FROM PUBLIC;
CREATE TRIGGER analytics_tenant_authorities_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.analytics_tenant_authorities
FOR EACH ROW EXECUTE FUNCTION platform.analytics_guard_tenant_authority_v1();
ALTER TABLE platform.analytics_tenant_authorities
    ENABLE ALWAYS TRIGGER analytics_tenant_authorities_guard;
-- +goose StatementEnd

