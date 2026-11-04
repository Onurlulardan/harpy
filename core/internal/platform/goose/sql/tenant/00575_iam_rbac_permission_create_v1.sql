-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_rbac_permission_create_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_match_kind text,
    in_resource_id uuid,
    in_action_id uuid,
    in_protected boolean,
    in_actor_id uuid,
    in_required_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_rbac_permission_create_v1(name, uuid, bytea, uuid, text, uuid, uuid, boolean, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_permission_create_v1(name, uuid, bytea, uuid, text, uuid, uuid, boolean, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_permission_create_v1(name, uuid, bytea, uuid, text, uuid, uuid, boolean, uuid, uuid, bigint, bigint);
