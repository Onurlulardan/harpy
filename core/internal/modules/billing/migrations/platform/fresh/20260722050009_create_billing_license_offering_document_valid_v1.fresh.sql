-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_offering_document_valid_v1(
    p_snapshot_id uuid, p_offering_version_id uuid,
    p_revision bigint, p_effective_at timestamptz,
    p_document jsonb
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_snapshot platform.billing_license_snapshots%ROWTYPE;
BEGIN
    SELECT snapshot.* INTO v_snapshot
    FROM platform.billing_license_snapshots AS snapshot
    WHERE snapshot.snapshot_id = p_snapshot_id;
    IF NOT FOUND
       OR pg_catalog.jsonb_typeof(p_document) <> 'object'
       OR (SELECT pg_catalog.count(*)
           FROM pg_catalog.jsonb_object_keys(p_document)) <> 13
       OR p_document ->> 'schema' <>
            'harpy-license-derived-offering.v1'
       OR p_document ->> 'offering_version_id' <>
            p_offering_version_id::text
       OR (p_document ->> 'revision')::bigint <> p_revision
       OR (p_document ->> 'effective_at')::timestamptz <> p_effective_at
       OR p_document ->> 'license_id' <> v_snapshot.license_id::text
       OR p_document ->> 'licensee_organization_id' <>
            v_snapshot.licensee_organization_id::text
       OR p_document ->> 'environment_id' <> v_snapshot.environment_id::text
       OR p_document ->> 'issuer' <> v_snapshot.issuer
       OR (p_document ->> 'snapshot_sequence')::bigint <> v_snapshot.sequence
       OR p_document ->> 'snapshot_sha256' <> v_snapshot.payload_sha256
       OR p_document ->> 'release_channel' <> v_snapshot.release_channel
       OR pg_catalog.jsonb_typeof(p_document -> 'booleans') <> 'array'
       OR pg_catalog.jsonb_array_length(p_document -> 'booleans') <> 5
       OR pg_catalog.jsonb_typeof(p_document -> 'limits') <> 'array'
       OR pg_catalog.jsonb_array_length(p_document -> 'limits') <> 5 THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.billing_license_snapshot_projections AS projection
        WHERE projection.snapshot_id = p_snapshot_id
          AND projection.value_type = 'boolean'
          AND NOT EXISTS (
              SELECT 1
              FROM pg_catalog.jsonb_to_recordset(
                  p_document -> 'booleans'
              ) AS grant_row(key text, value boolean)
              WHERE grant_row.key = projection.entitlement_key
                AND grant_row.value = projection.boolean_value
          )
    ) OR EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_to_recordset(
            p_document -> 'booleans'
        ) AS grant_row(key text, value boolean)
        WHERE NOT EXISTS (
            SELECT 1
            FROM platform.billing_license_snapshot_projections AS projection
            WHERE projection.snapshot_id = p_snapshot_id
              AND projection.value_type = 'boolean'
              AND projection.entitlement_key = grant_row.key
              AND projection.boolean_value = grant_row.value
        )
    ) THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.billing_license_snapshot_projections AS projection
        WHERE projection.snapshot_id = p_snapshot_id
          AND projection.value_type = 'integer'
          AND NOT EXISTS (
              SELECT 1
              FROM pg_catalog.jsonb_to_recordset(
                  p_document -> 'limits'
              ) AS grant_row(key text, value bigint, unlimited boolean)
              WHERE grant_row.key = projection.entitlement_key
                AND grant_row.unlimited = projection.unlimited
                AND grant_row.value = COALESCE(projection.integer_value, 0)
          )
    ) OR EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_to_recordset(
            p_document -> 'limits'
        ) AS grant_row(key text, value bigint, unlimited boolean)
        WHERE NOT EXISTS (
            SELECT 1
            FROM platform.billing_license_snapshot_projections AS projection
            WHERE projection.snapshot_id = p_snapshot_id
              AND projection.value_type = 'integer'
              AND projection.entitlement_key = grant_row.key
              AND projection.unlimited = grant_row.unlimited
              AND COALESCE(projection.integer_value, 0) = grant_row.value
        )
    ) THEN
        RETURN false;
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_offering_document_valid_v1(
    uuid, uuid, bigint, timestamptz, jsonb
) FROM PUBLIC;
