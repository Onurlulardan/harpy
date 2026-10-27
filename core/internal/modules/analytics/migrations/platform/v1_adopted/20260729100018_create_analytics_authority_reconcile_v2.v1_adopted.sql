-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.analytics_reconcile_current_tenant_authorities_v2()
RETURNS TABLE(reconciled_count integer, remaining_count integer)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog AS $function$
DECLARE
    v_tenant_id uuid;
    v_group_generation bigint;
    v_domain_generation bigint;
    v_reconciled integer;
    v_remaining integer;
BEGIN
    v_reconciled := 0;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR platform.tenancy_authority_reconcile_scope_assert_v1('analytics')
            IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics authority reconciliation is denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'forgecrud:analytics:tenant-authority-reconcile:v1', 0));
    FOR v_tenant_id, v_group_generation, v_domain_generation IN
        SELECT tenant.id, group_row.generation, domain_row.generation
        FROM platform.tenants AS tenant
        JOIN platform.cutover_groups AS group_row
          ON group_row.scope_kind = 'tenant'
         AND group_row.scope_id = tenant.id
         AND group_row.group_id = 'analytics'
        JOIN platform.cutover_group_members AS member_row
          ON member_row.scope_kind = group_row.scope_kind
         AND member_row.scope_id = group_row.scope_id
         AND member_row.group_id = group_row.group_id
         AND member_row.ordinal = 1
         AND member_row.domain = 'analytics'
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind = member_row.scope_kind
         AND domain_row.scope_id = member_row.scope_id
         AND domain_row.domain = member_row.domain
        WHERE tenant.deleted_at IS NULL
          AND tenant.status IN ('ACTIVE', 'TRIAL')
          AND group_row.state = 'v2' AND group_row.writer = 'v2'
          AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
          AND NOT EXISTS (
              SELECT 1 FROM platform.cutover_group_members AS extra
              WHERE extra.scope_kind = 'tenant'
                AND extra.scope_id = tenant.id
                AND extra.group_id = 'analytics'
                AND NOT (extra.ordinal = 1 AND extra.domain = 'analytics')
          )
          AND NOT EXISTS (
              SELECT 1
              FROM platform.analytics_tenant_authorities AS authority_row
              WHERE authority_row.tenant_id = tenant.id
                AND authority_row.authority_group_generation =
                    group_row.generation
                AND authority_row.authority_domain_generation =
                    domain_row.generation
                AND authority_row.authority_mode = 'runtime_v2'
          )
        ORDER BY tenant.id
    LOOP
        BEGIN
            PERFORM platform.analytics_register_tenant_authority_v1(
                pg_catalog.uuidv7(), v_tenant_id,
                v_group_generation, v_domain_generation,
                pg_catalog.transaction_timestamp()
            );
            v_reconciled := v_reconciled + 1;
        EXCEPTION WHEN SQLSTATE 'P6101' THEN
            IF NOT EXISTS (
                SELECT 1
                FROM platform.analytics_tenant_authorities AS authority_row
                JOIN platform.tenants AS tenant
                  ON tenant.id = authority_row.tenant_id
                JOIN platform.cutover_groups AS group_row
                  ON group_row.scope_kind = authority_row.scope_kind
                 AND group_row.scope_id = authority_row.scope_id
                 AND group_row.group_id = authority_row.authority_group_id
                JOIN platform.cutover_group_members AS member_row
                  ON member_row.scope_kind = authority_row.scope_kind
                 AND member_row.scope_id = authority_row.scope_id
                 AND member_row.group_id = authority_row.authority_group_id
                 AND member_row.ordinal = 1
                 AND member_row.domain = authority_row.authority_domain
                JOIN platform.domain_cutovers AS domain_row
                  ON domain_row.scope_kind = authority_row.scope_kind
                 AND domain_row.scope_id = authority_row.scope_id
                 AND domain_row.domain = authority_row.authority_domain
                WHERE authority_row.tenant_id = v_tenant_id
                  AND tenant.deleted_at IS NULL
                  AND tenant.status IN ('ACTIVE', 'TRIAL')
                  AND authority_row.scope_kind = 'tenant'
                  AND authority_row.scope_id = v_tenant_id
                  AND authority_row.authority_group_id = 'analytics'
                  AND authority_row.authority_domain = 'analytics'
                  AND authority_row.authority_group_generation =
                      v_group_generation
                  AND authority_row.authority_domain_generation =
                      v_domain_generation
                  AND authority_row.authority_mode = 'runtime_v2'
                  AND group_row.generation = v_group_generation
                  AND domain_row.generation = v_domain_generation
                  AND group_row.state = 'v2' AND group_row.writer = 'v2'
                  AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
                  AND NOT EXISTS (
                      SELECT 1
                      FROM platform.cutover_group_members AS extra
                      WHERE extra.scope_kind = 'tenant'
                        AND extra.scope_id = v_tenant_id
                        AND extra.group_id = 'analytics'
                        AND NOT (
                            extra.ordinal = 1 AND extra.domain = 'analytics'
                        )
                  )
            ) THEN
                RAISE;
            END IF;
        END;
    END LOOP;
    SELECT pg_catalog.count(*)::integer INTO v_remaining
    FROM platform.tenants AS tenant
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'tenant'
     AND group_row.scope_id = tenant.id
     AND group_row.group_id = 'analytics'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = 'tenant'
     AND domain_row.scope_id = tenant.id
     AND domain_row.domain = 'analytics'
    WHERE tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND group_row.state = 'v2' AND group_row.writer = 'v2'
      AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
      AND (
        NOT EXISTS (
            SELECT 1 FROM platform.cutover_group_members AS only_member
            WHERE only_member.scope_kind = 'tenant'
              AND only_member.scope_id = tenant.id
              AND only_member.group_id = 'analytics'
              AND only_member.ordinal = 1
              AND only_member.domain = 'analytics'
              AND NOT EXISTS (
                  SELECT 1 FROM platform.cutover_group_members AS extra
                  WHERE extra.scope_kind = 'tenant'
                    AND extra.scope_id = tenant.id
                    AND extra.group_id = 'analytics'
                    AND NOT (extra.ordinal = 1 AND extra.domain = 'analytics')
              )
        )
        OR NOT EXISTS (
            SELECT 1
            FROM platform.analytics_tenant_authorities AS authority_row
            WHERE authority_row.tenant_id = tenant.id
              AND authority_row.authority_group_generation =
                  group_row.generation
              AND authority_row.authority_domain_generation =
                  domain_row.generation
              AND authority_row.authority_mode = 'runtime_v2'
        )
      );
    IF v_remaining IS DISTINCT FROM 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6101',
            MESSAGE = 'analytics authority reconciliation is incomplete';
    END IF;
    RETURN QUERY SELECT v_reconciled, v_remaining;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.analytics_reconcile_current_tenant_authorities_v2() FROM PUBLIC;
