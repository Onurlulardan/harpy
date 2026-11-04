-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_get_signing_bundle_v1(
    in_environment_id text,
    in_generation bigint
)
RETURNS TABLE (
    environment_id text,
    generation bigint,
    bundle_sha256 text,
    bundle_bytes bytea,
    created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        0::bigint,
        ''::text,
        ''::bytea,
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_get_signing_bundle_v1(text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_signing_bundle_v1(text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_get_signing_bundle_v1(text, bigint);
