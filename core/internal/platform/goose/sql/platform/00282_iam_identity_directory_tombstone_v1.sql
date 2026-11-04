-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_identity_directory_tombstone_v1(
    in_tenant_id uuid,
    in_group_generation bigint,
    in_generation bigint,
    in_id uuid,
    in_revision bigint
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT 0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_identity_directory_tombstone_v1(uuid, bigint, bigint, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_identity_directory_tombstone_v1(uuid, bigint, bigint, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_identity_directory_tombstone_v1(uuid, bigint, bigint, uuid, bigint);
