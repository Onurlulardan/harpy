-- +goose Up
CREATE FUNCTION iam_initial_password_complete_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_selector bytea,
    in_secret_hmac bytea,
    in_csrf_hmac bytea,
    in_expected_continuation_revision bigint,
    in_user_id uuid,
    in_credential_id uuid,
    in_expected_credential_version bigint,
    in_expected_credential_revision bigint,
    in_salt bytea,
    in_tag bytea,
    in_request_sha bytea
)
RETURNS TABLE (
    user_id uuid,
    access_mode_id uuid,
    user_revision bigint,
    token_version bigint,
    authorization_version bigint,
    access_mode_revision bigint,
    credential_version bigint,
    credential_revision bigint,
    continuation_revision bigint,
    installation_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_user_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_initial_password_complete_v1(name, uuid, bytea, bytea, bytea, bigint, uuid, uuid, bigint, bigint, bytea, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_initial_password_complete_v1(name, uuid, bytea, bytea, bytea, bigint, uuid, uuid, bigint, bigint, bytea, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_initial_password_complete_v1(name, uuid, bytea, bytea, bytea, bigint, uuid, uuid, bigint, bigint, bytea, bytea, bytea);
