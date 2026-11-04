-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_create_v2(
    in_schema_name name,
    in_tenant_id uuid,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_identifier_id uuid,
    in_access_mode_id uuid,
    in_default_assignment_id uuid,
    in_schema_id uuid,
    in_values bytea,
    in_profile_sha bytea,
    in_display text,
    in_normalized text,
    in_email_sha bytea,
    in_lookup_key_id text,
    in_directory_hmac bytea,
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
        in_id,
        0::bigint,
        0::bigint,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_user_create_v2(name, uuid, uuid, bytea, uuid, uuid, uuid, uuid, uuid, bytea, bytea, text, text, bytea, text, bytea, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_user_create_v2(name, uuid, uuid, bytea, uuid, uuid, uuid, uuid, uuid, bytea, bytea, text, text, bytea, text, bytea, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_user_create_v2(name, uuid, uuid, bytea, uuid, uuid, uuid, uuid, uuid, bytea, bytea, text, text, bytea, text, bytea, uuid, uuid, bigint, bigint);
