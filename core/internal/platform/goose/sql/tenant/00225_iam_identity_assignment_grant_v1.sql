-- +goose Up
CREATE FUNCTION iam_identity_assignment_grant_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_user_id uuid,
    in_role_id uuid,
    in_scope_kind text,
    in_scope_id uuid,
    in_organization_id uuid,
    in_membership_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
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
        in_id,
        0::bigint,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_identity_assignment_grant_v1(name, uuid, bytea, uuid, uuid, uuid, text, uuid, uuid, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_assignment_grant_v1(name, uuid, bytea, uuid, uuid, uuid, text, uuid, uuid, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_assignment_grant_v1(name, uuid, bytea, uuid, uuid, uuid, text, uuid, uuid, uuid, uuid, uuid, bigint, bigint);
