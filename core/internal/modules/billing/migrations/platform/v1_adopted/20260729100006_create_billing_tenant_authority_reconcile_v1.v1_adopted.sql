-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_tenant_binding_v2() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR SESSION_USER NOT IN ('harpy_v2_runtime', 'harpy_adopter',
           'harpy_migrator', 'harpy_cutover_finalizer') THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'billing tenant binding session is invalid';
    END IF;
    IF TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'billing tenant authority binding is immutable';
    END IF;
    PERFORM 1
    FROM platform.tenants AS tenant
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'tenant' AND group_row.scope_id = tenant.id
     AND group_row.group_id = 'billing_forms'
    JOIN platform.cutover_group_members AS billing_member
      ON billing_member.scope_kind = group_row.scope_kind
     AND billing_member.scope_id = group_row.scope_id
     AND billing_member.group_id = group_row.group_id
     AND billing_member.ordinal = 1 AND billing_member.domain = 'billing'
    JOIN platform.cutover_group_members AS forms_member
      ON forms_member.scope_kind = group_row.scope_kind
     AND forms_member.scope_id = group_row.scope_id
     AND forms_member.group_id = group_row.group_id
     AND forms_member.ordinal = 2 AND forms_member.domain = 'forms'
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = group_row.scope_kind
     AND billing_row.scope_id = group_row.scope_id
     AND billing_row.domain = billing_member.domain
    JOIN platform.domain_cutovers AS forms_row
      ON forms_row.scope_kind = group_row.scope_kind
     AND forms_row.scope_id = group_row.scope_id
     AND forms_row.domain = forms_member.domain
    WHERE tenant.id = NEW.tenant_id AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND NEW.scope_kind = 'tenant' AND NEW.scope_id = tenant.id
      AND NEW.authority_group_id = 'billing_forms'
      AND NEW.billing_domain = 'billing' AND NEW.forms_domain = 'forms'
      AND NEW.authority_group_generation = group_row.generation
      AND NEW.billing_generation = billing_row.generation
      AND NEW.forms_generation = forms_row.generation
      AND group_row.state = billing_row.state
      AND group_row.state = forms_row.state
      AND group_row.writer = billing_row.writer
      AND group_row.writer = forms_row.writer
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS extra
          WHERE extra.scope_kind = group_row.scope_kind
            AND extra.scope_id = group_row.scope_id
            AND extra.group_id = group_row.group_id
            AND NOT ((extra.ordinal = 1 AND extra.domain = 'billing')
                OR (extra.ordinal = 2 AND extra.domain = 'forms'))
      )
    FOR SHARE OF tenant, group_row, billing_member, forms_member,
        billing_row, forms_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721', MESSAGE = 'billing tenant authority binding is stale';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_guard_tenant_binding_v2() FROM PUBLIC;
CREATE OR REPLACE TRIGGER billing_tenant_authorities_guard BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_tenant_authorities FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_binding_v2();
ALTER TABLE platform.billing_tenant_authorities ENABLE ALWAYS TRIGGER billing_tenant_authorities_guard;
CREATE FUNCTION platform.billing_reconcile_current_tenant_authorities_v1()
RETURNS TABLE(reconciled_count integer, remaining_count integer)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_tenant_id uuid;
    v_group_generation bigint;
    v_billing_generation bigint;
    v_forms_generation bigint;
BEGIN
    reconciled_count := 0;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true) IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'billing tenant authority reconciliation is denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'forgecrud:billing:tenant-authority-reconcile:v1', 0));
    FOR v_tenant_id, v_group_generation, v_billing_generation,
        v_forms_generation IN
        SELECT tenant.id, group_row.generation, billing_row.generation,
               forms_row.generation
        FROM platform.tenants AS tenant
        JOIN platform.cutover_groups AS group_row
          ON group_row.scope_kind = 'tenant' AND group_row.scope_id = tenant.id
         AND group_row.group_id = 'billing_forms'
        JOIN platform.cutover_group_members AS billing_member
          ON billing_member.scope_kind = group_row.scope_kind
         AND billing_member.scope_id = group_row.scope_id
         AND billing_member.group_id = group_row.group_id
         AND billing_member.ordinal = 1 AND billing_member.domain = 'billing'
        JOIN platform.cutover_group_members AS forms_member
          ON forms_member.scope_kind = group_row.scope_kind
         AND forms_member.scope_id = group_row.scope_id
         AND forms_member.group_id = group_row.group_id
         AND forms_member.ordinal = 2 AND forms_member.domain = 'forms'
        JOIN platform.domain_cutovers AS billing_row
          ON billing_row.scope_kind = 'tenant' AND billing_row.scope_id = tenant.id
         AND billing_row.domain = 'billing'
        JOIN platform.domain_cutovers AS forms_row
          ON forms_row.scope_kind = 'tenant' AND forms_row.scope_id = tenant.id
         AND forms_row.domain = 'forms'
        WHERE tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE', 'TRIAL')
          AND group_row.state = 'v2' AND group_row.writer = 'v2'
          AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
          AND forms_row.state = 'v2' AND forms_row.writer = 'v2'
          AND NOT EXISTS (
              SELECT 1 FROM platform.cutover_group_members AS extra
              WHERE extra.scope_kind = 'tenant' AND extra.scope_id = tenant.id
                AND extra.group_id = 'billing_forms'
                AND NOT ((extra.ordinal = 1 AND extra.domain = 'billing')
                    OR (extra.ordinal = 2 AND extra.domain = 'forms'))
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_tenant_authorities AS authority
              WHERE authority.tenant_id = tenant.id
                AND authority.authority_group_generation = group_row.generation
                AND authority.billing_generation = billing_row.generation
                AND authority.forms_generation = forms_row.generation
          )
        ORDER BY tenant.id
        FOR SHARE OF tenant, group_row, billing_member, forms_member,
            billing_row, forms_row
    LOOP
        BEGIN
            PERFORM platform.billing_register_tenant_authority_v1(
                pg_catalog.uuidv7(), v_tenant_id, v_group_generation,
                v_billing_generation, v_forms_generation,
                pg_catalog.transaction_timestamp()
            );
            reconciled_count := reconciled_count + 1;
        EXCEPTION WHEN SQLSTATE '23505' THEN
            IF NOT EXISTS (
                SELECT 1 FROM platform.billing_tenant_authorities AS authority
                WHERE authority.tenant_id = v_tenant_id
                  AND authority.authority_group_generation = v_group_generation
                  AND authority.billing_generation = v_billing_generation
                  AND authority.forms_generation = v_forms_generation
            ) THEN
                RAISE;
            END IF;
        END;
    END LOOP;
    SELECT pg_catalog.count(*)::integer INTO remaining_count
    FROM platform.tenants AS tenant
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'tenant' AND group_row.scope_id = tenant.id
     AND group_row.group_id = 'billing_forms'
    JOIN platform.cutover_group_members AS billing_member
      ON billing_member.scope_kind = group_row.scope_kind
     AND billing_member.scope_id = group_row.scope_id
     AND billing_member.group_id = group_row.group_id
     AND billing_member.ordinal = 1 AND billing_member.domain = 'billing'
    JOIN platform.cutover_group_members AS forms_member
      ON forms_member.scope_kind = group_row.scope_kind
     AND forms_member.scope_id = group_row.scope_id
     AND forms_member.group_id = group_row.group_id
     AND forms_member.ordinal = 2 AND forms_member.domain = 'forms'
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = 'tenant' AND billing_row.scope_id = tenant.id
     AND billing_row.domain = 'billing'
    JOIN platform.domain_cutovers AS forms_row
      ON forms_row.scope_kind = 'tenant' AND forms_row.scope_id = tenant.id
     AND forms_row.domain = 'forms'
    WHERE tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND group_row.state = 'v2' AND group_row.writer = 'v2'
      AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
      AND forms_row.state = 'v2' AND forms_row.writer = 'v2'
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS extra
          WHERE extra.scope_kind = 'tenant' AND extra.scope_id = tenant.id
            AND extra.group_id = 'billing_forms'
            AND NOT ((extra.ordinal = 1 AND extra.domain = 'billing')
                OR (extra.ordinal = 2 AND extra.domain = 'forms'))
      )
      AND NOT EXISTS (
          SELECT 1 FROM platform.billing_tenant_authorities AS authority
          WHERE authority.tenant_id = tenant.id
            AND authority.authority_group_generation = group_row.generation
            AND authority.billing_generation = billing_row.generation
            AND authority.forms_generation = forms_row.generation
      );
    IF remaining_count IS DISTINCT FROM 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721', MESSAGE = 'billing tenant authority reconciliation is incomplete';
    END IF;
    RETURN QUERY SELECT reconciled_count, remaining_count;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.billing_reconcile_current_tenant_authorities_v1()
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_reconcile_current_tenant_authorities_v1() TO harpy_runtime;
