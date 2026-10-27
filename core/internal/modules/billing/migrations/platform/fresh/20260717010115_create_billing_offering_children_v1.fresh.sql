-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_add_offering_price_v1(
    p_authority_id uuid, p_offering_version_id uuid, p_price_id uuid,
    p_cadence text, p_minor_units bigint, p_currency text,
    p_currency_registry_version bigint, p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    selected_start timestamptz;
    selected_end timestamptz;
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    SELECT effective_start, effective_end INTO selected_start, selected_end
    FROM platform.billing_offering_versions
    WHERE authority_id = p_authority_id AND id = p_offering_version_id
      AND publication_state = 'published'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'billing offering version is missing';
    END IF;
    INSERT INTO platform.billing_price_versions (
        id, authority_id, offering_version_id, cadence, minor_units,
        currency, currency_registry_version, currency_eligible,
        effective_start, effective_end, money_policy_version, published_at
    ) VALUES (
        p_price_id, p_authority_id, p_offering_version_id, p_cadence,
        p_minor_units, p_currency::character(3), p_currency_registry_version,
        true, selected_start, selected_end, 'fc_money_v1', p_occurred_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_add_offering_price_v1(
    uuid,uuid,uuid,text,bigint,text,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_add_offering_price_v1(
    uuid,uuid,uuid,text,bigint,text,bigint,timestamptz
) TO harpy_runtime;

CREATE FUNCTION platform.billing_add_offering_grant_v1(
    p_authority_id uuid, p_offering_version_id uuid,
    p_entitlement_key text, p_value_type text, p_boolean_value boolean,
    p_integer_value bigint, p_unlimited boolean, p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    INSERT INTO platform.billing_offering_entitlements (
        authority_id, offering_version_id, entitlement_definition_id,
        value_type, definition_allows_unlimited, boolean_value,
        integer_value, unlimited, created_at
    )
    SELECT p_authority_id, p_offering_version_id, definition.id,
           definition.value_type, definition.unlimited_allowed,
           p_boolean_value, p_integer_value, p_unlimited, p_occurred_at
    FROM platform.billing_entitlement_definitions AS definition
    JOIN platform.billing_offering_versions AS offering_version
      ON offering_version.authority_id = definition.authority_id
     AND offering_version.id = p_offering_version_id
    WHERE definition.authority_id = p_authority_id
      AND definition.entitlement_key = p_entitlement_key
      AND definition.value_type = p_value_type;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'billing entitlement definition is missing';
    END IF;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_add_offering_grant_v1(
    uuid,uuid,text,text,boolean,bigint,boolean,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_add_offering_grant_v1(
    uuid,uuid,text,text,boolean,bigint,boolean,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd

