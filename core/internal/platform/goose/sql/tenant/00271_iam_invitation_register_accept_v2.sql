-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_register_accept_v2(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_expected_revision bigint,
    in_id_2 uuid,
    in_revision bigint,
    in_selector bytea,
    in_hash_key_id text,
    in_secret_hmac bytea,
    in_binding_sha256 bytea,
    in_accepted_user_id uuid,
    in_accepted_identifier_id uuid,
    in_access_mode_id uuid,
    in_role_assignment_id uuid,
    in_organization_membership_id uuid,
    in_schema_id uuid,
    in_values bytea,
    in_sha256 bytea,
    in_salt bytea,
    in_tag bytea,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    code text,
    revision bigint,
    authorization_clock bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        in_revision,
        0::bigint,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_register_accept_v2(name, uuid, bytea, uuid, bigint, uuid, bigint, bytea, text, bytea, bytea, uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bytea, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_register_accept_v2(name, uuid, bytea, uuid, bigint, uuid, bigint, bytea, text, bytea, bytea, uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bytea, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_register_accept_v2(name, uuid, bytea, uuid, bigint, uuid, bigint, bytea, text, bytea, bytea, uuid, uuid, uuid, uuid, uuid, uuid, bytea, bytea, bytea, bytea, bigint);
