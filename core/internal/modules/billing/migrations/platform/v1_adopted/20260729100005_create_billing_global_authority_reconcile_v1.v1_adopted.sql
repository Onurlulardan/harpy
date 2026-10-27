-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_global_binding_v2()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR SESSION_USER NOT IN (
           'harpy_v2_runtime', 'harpy_adopter', 'harpy_migrator',
           'harpy_cutover_finalizer'
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
     AND member_row.ordinal = 1 AND member_row.domain = 'billing'
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = member_row.scope_kind
     AND domain_row.scope_id = member_row.scope_id
     AND domain_row.domain = member_row.domain
    WHERE NEW.scope_kind = 'global'
      AND NEW.scope_id = environment.environment_id
      AND NEW.authority_group_id = 'billing'
      AND NEW.authority_domain = 'billing'
      AND NEW.authority_group_generation = group_row.generation
      AND NEW.authority_domain_generation = domain_row.generation
      AND group_row.state = domain_row.state
      AND group_row.writer = domain_row.writer
      AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_group_members AS extra
          WHERE extra.scope_kind = group_row.scope_kind
            AND extra.scope_id = group_row.scope_id
            AND extra.group_id = group_row.group_id
            AND NOT (extra.ordinal = 1 AND extra.domain = 'billing')
      )
    FOR SHARE OF group_row, member_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing global authority binding is stale';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_guard_global_binding_v2() FROM PUBLIC;
CREATE OR REPLACE TRIGGER billing_global_authorities_guard BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_global_authorities FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_global_binding_v2();
ALTER TABLE platform.billing_global_authorities ENABLE ALWAYS TRIGGER billing_global_authorities_guard;

CREATE FUNCTION platform.billing_reconcile_current_global_authority_v1()
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
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing global authority reconciliation is denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'harpy:billing:global-authority-reconcile:v1', 0
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
    platform.billing_reconcile_current_global_authority_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.billing_reconcile_current_global_authority_v1()
    TO harpy_runtime;
