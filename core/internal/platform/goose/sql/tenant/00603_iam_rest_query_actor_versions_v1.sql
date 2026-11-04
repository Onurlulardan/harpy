-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_rest_query_actor_versions_v1(
    in_schema_name name,
    in_actor_id uuid
)
RETURNS TABLE (
    token_version bigint,
    authorization_version bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        0::bigint,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_rest_query_actor_versions_v1(name, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rest_query_actor_versions_v1(name, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rest_query_actor_versions_v1(name, uuid);
