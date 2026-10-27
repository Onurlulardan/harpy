-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_offering_insert_v1(
    p_authority_id uuid, p_operation_id uuid,
    p_offering_id uuid, p_offering_version_id uuid,
    p_revision bigint, p_effective_at timestamptz,
    p_snapshot_id uuid
)
RETURNS integer
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_snapshot platform.billing_license_snapshots%ROWTYPE;
    v_metadata jsonb;
    v_count integer;
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    SELECT snapshot.* INTO v_snapshot
    FROM platform.billing_license_snapshots AS snapshot
    WHERE snapshot.snapshot_id = p_snapshot_id;
    IF NOT FOUND OR p_revision < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing license offering input is invalid';
    END IF;
    v_metadata := pg_catalog.jsonb_build_object(
        'name', 'Harpy License',
        'description', 'License-derived on-premise base offering',
        'sort_order', 0,
        'highlight_label', '',
        'color', '',
        'billing_interval', 'monthly'
    );
    INSERT INTO platform.billing_offerings (
        id, authority_id, offering_key, lifecycle,
        created_at, retired_at, revision
    ) VALUES (
        p_offering_id, p_authority_id,
        'license.' || v_snapshot.license_id::text || '.'
            || v_snapshot.environment_id::text,
        'active', p_effective_at, NULL, p_revision
    );
    INSERT INTO platform.billing_offering_versions (
        id, authority_id, offering_id, version,
        is_base, is_public, available, trial_days, grace_days,
        expected_price_count, expected_grant_count,
        effective_start, effective_end, presentation_metadata,
        presentation_sha256, publication_state,
        publication_operation_id, published_at
    ) VALUES (
        p_offering_version_id, p_authority_id, p_offering_id, p_revision,
        true, false, false, 0, 0, 1, 13,
        p_effective_at, NULL, v_metadata,
        pg_catalog.sha256(pg_catalog.convert_to(v_metadata::text, 'UTF8')),
        'published', p_operation_id, p_effective_at
    );
    INSERT INTO platform.billing_price_versions (
        id, authority_id, offering_version_id, cadence,
        minor_units, currency, currency_registry_version,
        currency_eligible, effective_start, effective_end,
        money_policy_version, published_at
    ) VALUES (
        pg_catalog.gen_random_uuid(), p_authority_id,
        p_offering_version_id, 'one_time', 0, 'USD', 1,
        true, p_effective_at, NULL, 'fc_money_v1', p_effective_at
    );
    INSERT INTO platform.billing_offering_entitlements (
        authority_id, offering_version_id, entitlement_definition_id,
        value_type, definition_allows_unlimited, boolean_value,
        integer_value, unlimited, created_at
    )
    SELECT p_authority_id, p_offering_version_id, definition.id,
           definition.value_type, definition.unlimited_allowed,
           projection.boolean_value, projection.integer_value,
           projection.unlimited, p_effective_at
    FROM platform.billing_license_snapshot_projections AS projection
    JOIN platform.billing_entitlement_definitions AS definition
      ON definition.entitlement_key = projection.entitlement_key
     AND definition.value_type = projection.value_type
     AND (NOT projection.unlimited OR definition.unlimited_allowed)
    WHERE projection.snapshot_id = p_snapshot_id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count <> 13 THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'billing license offering grants are incomplete';
    END IF;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_offering_insert_v1(
    uuid, uuid, uuid, uuid, bigint, timestamptz, uuid
) FROM PUBLIC;

