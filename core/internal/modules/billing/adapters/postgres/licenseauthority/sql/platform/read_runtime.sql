SELECT COALESCE(
    pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object(
            'license_id', runtime.license_id,
            'environment_id', runtime.environment_id,
            'snapshot_sequence', runtime.snapshot_sequence,
            'snapshot_sha256', runtime.payload_sha256,
            'state', runtime.runtime_state,
            'entitlement_key', runtime.entitlement_key,
            'value_type', runtime.value_type,
            'boolean_value', runtime.boolean_value,
            'integer_value', runtime.integer_value,
            'unlimited', runtime.unlimited,
            'channel', runtime.release_channel,
            'database_now', runtime.database_now
        ) ORDER BY runtime.entitlement_key
    ),
    '[]'::jsonb
)::text
FROM platform.billing_license_runtime_entitlements_v1($1) AS runtime;
