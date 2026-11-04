-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_relations_replace_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_command_key text,
    in_request_sha256 bytea,
    in_user_id uuid,
    in_expected_revision bigint,
    in_patch_profile boolean,
    in_profile_schema_id uuid,
    in_profile_document bytea,
    in_profile_sha256 bytea,
    in_replace boolean,
    in_organization_ids uuid[],
    in_membership_ids uuid[],
    in_replace_2 boolean,
    in_role_ids uuid[],
    in_assignment_ids uuid[],
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
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_user_relations_replace_v1(name, uuid, text, bytea, uuid, bigint, boolean, uuid, bytea, bytea, boolean, uuid[], uuid[], boolean, uuid[], uuid[], uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_user_relations_replace_v1(name, uuid, text, bytea, uuid, bigint, boolean, uuid, bytea, bytea, boolean, uuid[], uuid[], boolean, uuid[], uuid[], uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_user_relations_replace_v1(name, uuid, text, bytea, uuid, bigint, boolean, uuid, bytea, bytea, boolean, uuid[], uuid[], boolean, uuid[], uuid[], uuid, uuid, bigint, bigint);
