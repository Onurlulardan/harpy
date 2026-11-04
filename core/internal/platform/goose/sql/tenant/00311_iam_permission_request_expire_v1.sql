-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_expire_v1(
    in_schema_name name,
    in_group_id uuid,
    in_expected_revision bigint,
    in_expected_state text
)
RETURNS TABLE (
    group_id uuid,
    group_state text,
    group_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_group_id,
        ''::text,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_expire_v1(name, uuid, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_request_expire_v1(name, uuid, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_request_expire_v1(name, uuid, bigint, text);
