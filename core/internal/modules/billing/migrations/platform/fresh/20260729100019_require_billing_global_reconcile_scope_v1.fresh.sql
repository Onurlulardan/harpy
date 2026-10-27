-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_reconcile_current_global_authority_v2()
RETURNS TABLE(reconciled_count integer, remaining_count integer)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_group_generation bigint;
    v_domain_generation bigint;
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
            MESSAGE = 'billing global authority reconciliation is denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'harpy:billing:global-authority-reconcile:v2', 0
        )
    );
    SELECT group_row.generation, domain_row.generation
    INTO STRICT v_group_generation, v_domain_generation
    FROM platform.environment_identity AS environment
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = 'global'
     AND group_row.scope_id = environment.environment_id
     AND group_row.group_id = 'billing'
    JOIN platform.cutover_group_members AS member_row
      ON member_row.scope_kind = group_row.scope_kind
     AND member_row.scope_id = group_row.scope_id
     AND member_row.group_id = group_row.group_id
     AND member_row.ordinal = 1 AND member_row.domain = 'billing'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = member_row.scope_kind
     AND domain_row.scope_id = member_row.scope_id
     AND domain_row.domain = member_row.domain
    WHERE group_row.state = 'v2' AND group_row.writer = 'v2'
      AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS extra
          WHERE extra.scope_kind = group_row.scope_kind
            AND extra.scope_id = group_row.scope_id
            AND extra.group_id = group_row.group_id
            AND NOT (extra.ordinal = 1 AND extra.domain = 'billing')
      )
    FOR SHARE OF group_row, member_row, domain_row;
    IF NOT EXISTS (
        SELECT 1 FROM platform.billing_global_authorities AS authority
        WHERE authority.authority_group_generation = v_group_generation
          AND authority.authority_domain_generation = v_domain_generation
    ) THEN
        BEGIN
            PERFORM platform.billing_register_global_authority_v1(
                pg_catalog.uuidv7(), v_group_generation, v_domain_generation,
                pg_catalog.transaction_timestamp()
            );
            reconciled_count := 1;
        EXCEPTION WHEN SQLSTATE '23505' THEN
            IF NOT EXISTS (
                SELECT 1 FROM platform.billing_global_authorities AS authority
                WHERE authority.authority_group_generation = v_group_generation
                  AND authority.authority_domain_generation = v_domain_generation
            ) THEN
                RAISE;
            END IF;
        END;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.billing_global_authorities AS authority
        WHERE authority.authority_group_generation = v_group_generation
          AND authority.authority_domain_generation = v_domain_generation
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing global authority reconciliation is incomplete';
    END IF;
    RETURN QUERY SELECT reconciled_count, 0;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION
    platform.billing_reconcile_current_global_authority_v2() FROM PUBLIC;
