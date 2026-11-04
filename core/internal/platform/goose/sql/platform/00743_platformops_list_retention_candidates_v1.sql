-- +goose Up
CREATE FUNCTION platform.platformops_list_retention_candidates_v1(
    in_environment_id text
)
RETURNS TABLE (
    recovery_set_id text,
    manifest_sha256 text,
    protected_at timestamptz,
    legal_hold boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        pg_catalog.clock_timestamp(),
        false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_list_retention_candidates_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_list_retention_candidates_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_list_retention_candidates_v1(text);
