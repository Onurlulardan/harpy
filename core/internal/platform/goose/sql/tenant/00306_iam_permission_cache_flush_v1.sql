-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_cache_flush_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_actor_id uuid,
    in_token_version bigint,
    in_authorization_version bigint,
    in_platform_clock bigint
)
RETURNS TABLE (
    authorization_clock bigint,
    result_revision bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        0::bigint,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_cache_flush_v1(name, uuid, uuid, bytea, uuid, bigint, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_cache_flush_v1(name, uuid, uuid, bytea, uuid, bigint, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_cache_flush_v1(name, uuid, uuid, bytea, uuid, bigint, bigint, bigint);
