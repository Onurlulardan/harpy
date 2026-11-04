-- +goose Up
CREATE FUNCTION platform.platformauthority_commit_v1(
    in_operation_id text,
    in_now timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_commit_v1(text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_commit_v1(text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_commit_v1(text, timestamptz);
