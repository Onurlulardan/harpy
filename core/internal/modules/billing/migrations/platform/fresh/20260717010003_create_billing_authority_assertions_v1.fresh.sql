-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_assert_global_authority_v1(p_authority_id uuid)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    expected_writer text;
    expected_state text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing global authority owner is invalid';
    END IF;
    IF SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime' THEN
        expected_writer := 'v2'; expected_state := 'v2';
    ELSIF SESSION_USER IS NOT DISTINCT FROM 'harpy_adopter' THEN
        expected_writer := 'v1'; expected_state := 'shadow';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing global authority session is invalid';
    END IF;
    PERFORM 1
    FROM platform.billing_global_authorities AS authority_row
    JOIN platform.environment_identity AS environment
      ON environment.environment_id = authority_row.scope_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority_row.scope_kind
     AND group_row.scope_id = authority_row.scope_id
     AND group_row.group_id = authority_row.authority_group_id
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = authority_row.scope_kind
     AND domain_row.scope_id = authority_row.scope_id
     AND domain_row.domain = authority_row.authority_domain
    WHERE authority_row.authority_id = p_authority_id
      AND authority_row.authority_group_generation = group_row.generation
      AND authority_row.authority_domain_generation = domain_row.generation
      AND group_row.writer = expected_writer
      AND domain_row.writer = expected_writer
      AND group_row.state = expected_state
      AND domain_row.state = expected_state
    FOR SHARE OF authority_row, group_row, domain_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing global authority is not current';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_assert_global_authority_v1(uuid)
    FROM PUBLIC;

CREATE FUNCTION platform.billing_assert_tenant_authority_v1(p_authority_id uuid)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    expected_writer text;
    expected_state text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing tenant authority owner is invalid';
    END IF;
    IF SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime' THEN
        expected_writer := 'v2'; expected_state := 'v2';
    ELSIF SESSION_USER IS NOT DISTINCT FROM 'harpy_adopter' THEN
        expected_writer := 'v1'; expected_state := 'shadow';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing tenant authority session is invalid';
    END IF;
    PERFORM 1
    FROM platform.billing_tenant_authorities AS authority_row
    JOIN platform.tenants AS tenant ON tenant.id = authority_row.tenant_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority_row.scope_kind
     AND group_row.scope_id = authority_row.scope_id
     AND group_row.group_id = authority_row.authority_group_id
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = authority_row.scope_kind
     AND billing_row.scope_id = authority_row.scope_id
     AND billing_row.domain = authority_row.billing_domain
    JOIN platform.domain_cutovers AS forms_row
      ON forms_row.scope_kind = authority_row.scope_kind
     AND forms_row.scope_id = authority_row.scope_id
     AND forms_row.domain = authority_row.forms_domain
    WHERE authority_row.authority_id = p_authority_id
      AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND authority_row.authority_group_generation = group_row.generation
      AND authority_row.billing_generation = billing_row.generation
      AND authority_row.forms_generation = forms_row.generation
      AND group_row.writer = expected_writer
      AND billing_row.writer = expected_writer
      AND forms_row.writer = expected_writer
      AND group_row.state = expected_state
      AND billing_row.state = expected_state
      AND forms_row.state = expected_state
    FOR SHARE OF authority_row, tenant, group_row, billing_row, forms_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5721',
            MESSAGE = 'billing tenant authority is not current';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_assert_tenant_authority_v1(uuid)
    FROM PUBLIC;
-- +goose StatementEnd
