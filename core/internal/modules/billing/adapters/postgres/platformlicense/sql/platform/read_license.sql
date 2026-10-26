WITH runtime AS (
    SELECT *
    FROM platform.billing_license_runtime_entitlements_v1($1)
)
SELECT runtime.license_id,
       runtime.environment_id,
       runtime.snapshot_sequence,
       runtime.payload_sha256,
       runtime.runtime_state,
       runtime.release_channel,
       pg_catalog.jsonb_object_agg(
           runtime.entitlement_key,
           runtime.boolean_value
           ORDER BY runtime.entitlement_key
       ) FILTER (WHERE runtime.value_type = 'boolean')::text,
       pg_catalog.jsonb_object_agg(
           runtime.entitlement_key,
           pg_catalog.jsonb_build_object(
               'unlimited', runtime.unlimited,
               'value', COALESCE(runtime.integer_value, 0)
           )
           ORDER BY runtime.entitlement_key
       ) FILTER (WHERE runtime.value_type = 'integer')::text,
       runtime.database_now
FROM runtime
GROUP BY runtime.license_id,
         runtime.environment_id,
         runtime.snapshot_sequence,
         runtime.payload_sha256,
         runtime.runtime_state,
         runtime.release_channel,
         runtime.database_now;
