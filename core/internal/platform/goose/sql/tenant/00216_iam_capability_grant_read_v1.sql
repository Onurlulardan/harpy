-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_capability_grant_read_v1(
    in_schema_name name,
    in_grant_id uuid,
    in_grant_revision bigint
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT ''::bytea;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_capability_grant_read_v1(name, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_capability_grant_read_v1(name, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_capability_grant_read_v1(name, uuid, bigint);
