-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_projection_insert_v1(
    p_snapshot_id uuid, p_projection jsonb, p_created_at timestamptz
)
RETURNS integer
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_keys text;
    v_count integer;
BEGIN
    IF pg_catalog.jsonb_typeof(p_projection) <> 'array'
       OR pg_catalog.jsonb_array_length(p_projection) <> 10
       OR EXISTS (
           SELECT 1
           FROM pg_catalog.jsonb_array_elements(p_projection) AS item
           WHERE pg_catalog.jsonb_typeof(item) <> 'object'
              OR (SELECT pg_catalog.count(*)
                  FROM pg_catalog.jsonb_object_keys(item)) <> 5
              OR NOT item ? 'entitlement_key'
              OR NOT item ? 'value_type'
              OR NOT item ? 'boolean_value'
              OR NOT item ? 'integer_value'
              OR NOT item ? 'unlimited'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing license projection shape is invalid';
    END IF;
    SELECT pg_catalog.string_agg(
        grant_row.entitlement_key, ',' ORDER BY grant_row.entitlement_key
    )
    INTO v_keys
    FROM pg_catalog.jsonb_to_recordset(p_projection) AS grant_row(
        entitlement_key text, value_type text, boolean_value boolean,
        integer_value bigint, unlimited boolean
    );
    IF v_keys IS DISTINCT FROM
        'analytics.active_dashboards,analytics.dashboard_builder.enabled,' ||
        'api.access,api.active_endpoints,api.builder.enabled,documents.storage_bytes,' ||
        'iam.active_users,' ||
        'workflow.active_definitions,workflow.builder.enabled,workflow.enabled' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing license projection keys are invalid';
    END IF;
    INSERT INTO platform.billing_license_snapshot_projections (
        snapshot_id, entitlement_key, value_type, boolean_value,
        integer_value, unlimited, created_at
    )
    SELECT p_snapshot_id, grant_row.entitlement_key,
           grant_row.value_type, grant_row.boolean_value,
           grant_row.integer_value, grant_row.unlimited, p_created_at
    FROM pg_catalog.jsonb_to_recordset(p_projection) AS grant_row(
        entitlement_key text, value_type text, boolean_value boolean,
        integer_value bigint, unlimited boolean
    )
    JOIN platform.billing_entitlement_definitions AS definition
      ON definition.entitlement_key = grant_row.entitlement_key
     AND definition.value_type = grant_row.value_type
     AND (NOT grant_row.unlimited OR definition.unlimited_allowed);
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count <> 10 THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'billing license projection definitions are incomplete';
    END IF;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_projection_insert_v1(
    uuid, jsonb, timestamptz
) FROM PUBLIC;
