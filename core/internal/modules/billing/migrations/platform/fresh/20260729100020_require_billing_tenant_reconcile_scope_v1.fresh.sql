-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_reconcile_current_tenant_authorities_v2()
RETURNS TABLE(reconciled_count integer, remaining_count integer)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_tenant_id uuid;
    v_group_generation bigint;
    v_billing_generation bigint;
BEGIN
    reconciled_count := 0;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize'
       OR platform.tenancy_authority_reconcile_scope_assert_v1(
           'billing'
       ) IS DISTINCT FROM TRUE THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing tenant authority reconciliation is denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'harpy:billing:tenant-authority-reconcile:v2', 0));
    FOR v_tenant_id, v_group_generation, v_billing_generation IN
        SELECT tenant.id, group_row.generation, billing_row.generation
        FROM platform.tenants AS tenant
        JOIN platform.cutover_groups AS group_row
          ON group_row.scope_kind = 'tenant' AND group_row.scope_id = tenant.id
         AND group_row.group_id = 'billing_forms'
        JOIN platform.cutover_group_members AS billing_member
          ON billing_member.scope_kind = group_row.scope_kind
         AND billing_member.scope_id = group_row.scope_id
         AND billing_member.group_id = group_row.group_id
         AND billing_member.ordinal = 1 AND billing_member.domain = 'billing'
        JOIN platform.domain_cutovers AS billing_row
          ON billing_row.scope_kind = 'tenant' AND billing_row.scope_id = tenant.id
         AND billing_row.domain = 'billing'
        WHERE tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE', 'TRIAL')
          AND group_row.state = 'v2' AND group_row.writer = 'v2'
          AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
          AND NOT EXISTS (
              SELECT 1 FROM platform.cutover_group_members AS extra
              WHERE extra.scope_kind = 'tenant' AND extra.scope_id = tenant.id
                AND extra.group_id = 'billing_forms'
                AND NOT (extra.ordinal = 1 AND extra.domain = 'billing')
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_tenant_authorities AS authority
              WHERE authority.tenant_id = tenant.id
                AND authority.authority_group_generation = group_row.generation
                AND authority.billing_generation = billing_row.generation
          )
        ORDER BY tenant.id
        FOR SHARE OF tenant, group_row, billing_member, billing_row
    LOOP
        BEGIN
            PERFORM platform.billing_register_tenant_authority_v1(
                pg_catalog.uuidv7(), v_tenant_id, v_group_generation,
                v_billing_generation, pg_catalog.transaction_timestamp()
            );
            reconciled_count := reconciled_count + 1;
        EXCEPTION WHEN SQLSTATE '23505' THEN
            IF NOT EXISTS (
                SELECT 1 FROM platform.billing_tenant_authorities AS authority
                WHERE authority.tenant_id = v_tenant_id
                  AND authority.authority_group_generation = v_group_generation
                  AND authority.billing_generation = v_billing_generation
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
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = 'tenant' AND billing_row.scope_id = tenant.id
     AND billing_row.domain = 'billing'
    WHERE tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND group_row.state = 'v2' AND group_row.writer = 'v2'
      AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS extra
          WHERE extra.scope_kind = 'tenant' AND extra.scope_id = tenant.id
            AND extra.group_id = 'billing_forms'
            AND NOT (extra.ordinal = 1 AND extra.domain = 'billing')
      )
      AND NOT EXISTS (
          SELECT 1 FROM platform.billing_tenant_authorities AS authority
          WHERE authority.tenant_id = tenant.id
            AND authority.authority_group_generation = group_row.generation
            AND authority.billing_generation = billing_row.generation
      );
    IF remaining_count IS DISTINCT FROM 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing tenant authority reconciliation is incomplete';
    END IF;
    RETURN QUERY SELECT reconciled_count, remaining_count;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION
    platform.billing_reconcile_current_tenant_authorities_v2() FROM PUBLIC;
