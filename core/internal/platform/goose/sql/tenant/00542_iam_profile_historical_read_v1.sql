-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_profile_historical_read_v1(
    in_schema_name name,
    in_schema_id uuid
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT ''::bytea;
$$;

REVOKE ALL ON FUNCTION iam_profile_historical_read_v1(name, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_profile_historical_read_v1(name, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_profile_historical_read_v1(name, uuid);
