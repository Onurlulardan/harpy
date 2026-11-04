-- +goose Up
CREATE FUNCTION iam_platform_operator_grant_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_actor_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_expected_platform_clock bigint,
    in_purpose_sha256 bytea,
    in_auth_time timestamptz,
    in_acr text,
    in_amr_password boolean,
    in_amr_web_authn boolean,
    in_amrtotp boolean,
    in_target_user_id uuid,
    in_target_expected_authorization_version bigint,
    in_subject_id uuid,
    in_role_id uuid,
    in_subject_role_binding_id uuid,
    in_role_permission_binding_id uuid,
    in_capability text
)
RETURNS TABLE (
    target_user_id uuid,
    subject_id uuid,
    role_id uuid,
    subject_role_binding_id uuid,
    role_permission_binding_id uuid,
    target_authorization_version bigint,
    platform_clock bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_target_user_id,
        in_subject_id,
        in_role_id,
        in_subject_role_binding_id,
        in_role_permission_binding_id,
        0::bigint,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_platform_operator_grant_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bigint, uuid, uuid, uuid, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_platform_operator_grant_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bigint, uuid, uuid, uuid, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_platform_operator_grant_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bigint, uuid, uuid, uuid, uuid, text);
