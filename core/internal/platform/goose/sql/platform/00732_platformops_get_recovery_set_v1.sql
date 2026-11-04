-- +goose Up
CREATE FUNCTION platform.platformops_get_recovery_set_v1(
    in_recovery_set_id text
)
RETURNS TABLE (
    recovery_set_id text,
    operation_id text,
    environment_id text,
    sequence bigint,
    state text,
    manifest_sha256 text,
    manifest_bytes bytea,
    legal_hold boolean,
    database_system_id text,
    recovery_signing_generation bigint,
    recovery_trust_bundle_sha256 text,
    protected_at timestamptz,
    created_at timestamptz,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        NULL::text,
        ''::bytea,
        false,
        NULL::text,
        0::bigint,
        ''::text,
        NULL::timestamptz,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.platformops_get_recovery_set_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_recovery_set_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_get_recovery_set_v1(text);
