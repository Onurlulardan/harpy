-- +goose Up
CREATE FUNCTION iam_invitation_expire_v2(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_expected_revision bigint,
    in_actor_id uuid,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    state text,
    revision bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_id,
        ''::text,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_invitation_expire_v2(name, uuid, bytea, uuid, bigint, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_expire_v2(name, uuid, bytea, uuid, bigint, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_expire_v2(name, uuid, bytea, uuid, bigint, uuid, bigint);
