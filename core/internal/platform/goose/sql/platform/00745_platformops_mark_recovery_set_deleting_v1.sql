-- +goose Up
CREATE FUNCTION platform.platformops_mark_recovery_set_deleting_v1(
    in_recovery_set_id text,
    in_manifest_sha256 text,
    in_updated_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_mark_recovery_set_deleting_v1(text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_mark_recovery_set_deleting_v1(text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_mark_recovery_set_deleting_v1(text, text, timestamptz);
