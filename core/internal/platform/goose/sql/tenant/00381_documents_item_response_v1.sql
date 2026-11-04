-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION documents_item_response_v1(
    arg_1 uuid
)
RETURNS jsonb
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT '{}'::jsonb;
$$;

REVOKE ALL ON FUNCTION documents_item_response_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_item_response_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION documents_item_response_v1(uuid);
