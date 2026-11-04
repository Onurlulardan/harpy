-- +goose Up
CREATE FUNCTION platform.billing_license_runtime_entitlements_v1(
    p_environment_id uuid
)
RETURNS TABLE (
    license_id uuid,
    environment_id uuid,
    snapshot_sequence bigint,
    payload_sha256 text,
    runtime_state text,
    entitlement_key text,
    value_type text,
    boolean_value boolean,
    integer_value bigint,
    unlimited boolean,
    release_channel text,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        snapshot.license_id,
        snapshot.environment_id,
        snapshot.sequence,
        snapshot.payload_sha256,
        snapshot.status,
        entry.key,
        CASE pg_catalog.jsonb_typeof(entry.value)
            WHEN 'number' THEN 'integer'
            ELSE 'boolean'
        END,
        CASE
            WHEN pg_catalog.jsonb_typeof(entry.value) = 'boolean'
                THEN entry.value::text::boolean
            ELSE NULL
        END,
        CASE
            WHEN pg_catalog.jsonb_typeof(entry.value) = 'number'
                THEN pg_catalog.trunc(entry.value::text::numeric)::bigint
            ELSE NULL
        END,
        false,
        snapshot.channel,
        pg_catalog.clock_timestamp()
    FROM platform.billing_license_snapshots AS snapshot
    CROSS JOIN LATERAL pg_catalog.jsonb_each(
        COALESCE(snapshot.features, '{}'::jsonb) || COALESCE(snapshot.limits, '{}'::jsonb)
    ) AS entry(key, value)
    WHERE snapshot.environment_id = p_environment_id
      AND snapshot.sequence = (
          SELECT pg_catalog.max(latest.sequence)
          FROM platform.billing_license_snapshots AS latest
          WHERE latest.environment_id = p_environment_id
      )
      AND pg_catalog.jsonb_typeof(entry.value) IN ('boolean', 'number');
$$;

REVOKE ALL ON FUNCTION platform.billing_license_runtime_entitlements_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_runtime_entitlements_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_runtime_entitlements_v1(uuid);
