-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_global_binding_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR SESSION_USER NOT IN (
           'forgecrud_v2_runtime', 'forgecrud_adopter', 'forgecrud_migrator'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing global binding session is invalid';
    END IF;
    IF TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing global authority binding is immutable';
    END IF;
    PERFORM 1
    FROM platform.environment_identity AS environment
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'global'
     AND group_row.scope_id = environment.environment_id
     AND group_row.group_id = 'billing'
    JOIN platform.cutover_group_members AS member_row
      ON member_row.scope_kind = group_row.scope_kind
     AND member_row.scope_id = group_row.scope_id
     AND member_row.group_id = group_row.group_id
     AND member_row.domain = 'billing'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = member_row.scope_kind
     AND domain_row.scope_id = member_row.scope_id
     AND domain_row.domain = member_row.domain
    WHERE NEW.scope_id = environment.environment_id
      AND NEW.authority_group_generation = group_row.generation
      AND NEW.authority_domain_generation = domain_row.generation
      AND group_row.state = domain_row.state
      AND group_row.writer = domain_row.writer
    FOR SHARE OF group_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing global authority binding is stale';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_global_binding_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_tenant_binding_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR SESSION_USER NOT IN (
           'forgecrud_v2_runtime', 'forgecrud_adopter', 'forgecrud_migrator'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing tenant binding session is invalid';
    END IF;
    IF TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing tenant authority binding is immutable';
    END IF;
    PERFORM 1
    FROM platform.tenants AS tenant
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'tenant'
     AND group_row.scope_id = tenant.id
     AND group_row.group_id = 'billing_forms'
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = group_row.scope_kind
     AND billing_row.scope_id = group_row.scope_id
     AND billing_row.domain = 'billing'
    JOIN platform.domain_cutovers AS forms_row
      ON forms_row.scope_kind = group_row.scope_kind
     AND forms_row.scope_id = group_row.scope_id
     AND forms_row.domain = 'forms'
    WHERE tenant.id = NEW.tenant_id
      AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND NEW.authority_group_generation = group_row.generation
      AND NEW.billing_generation = billing_row.generation
      AND NEW.forms_generation = forms_row.generation
      AND group_row.state = billing_row.state
      AND group_row.state = forms_row.state
      AND group_row.writer = billing_row.writer
      AND group_row.writer = forms_row.writer
      AND EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS member_row
          WHERE member_row.scope_kind = group_row.scope_kind
            AND member_row.scope_id = group_row.scope_id
            AND member_row.group_id = group_row.group_id
            AND member_row.domain IN ('billing', 'forms')
          HAVING pg_catalog.count(*) = 2
      )
    FOR SHARE OF tenant, group_row, billing_row, forms_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing tenant authority binding is stale';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_tenant_binding_v1() FROM PUBLIC;

CREATE TRIGGER billing_global_authorities_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_global_authorities
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_binding_v1();
ALTER TABLE platform.billing_global_authorities
    ENABLE ALWAYS TRIGGER billing_global_authorities_guard;

CREATE TRIGGER billing_tenant_authorities_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_tenant_authorities
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_binding_v1();
ALTER TABLE platform.billing_tenant_authorities
    ENABLE ALWAYS TRIGGER billing_tenant_authorities_guard;
-- +goose StatementEnd
