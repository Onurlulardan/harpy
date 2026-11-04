-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_reauthentication_start_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_binding bytea,
    in_id uuid,
    in_user_id uuid,
    in_session_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_access_mode text,
    in_portal_id uuid,
    in_method text,
    in_authenticator_id uuid,
    in_hmac_key_id text,
    in_challenge_hmac bytea,
    in_context_sha bytea,
    in_purpose_sha bytea,
    in_action_policy_version bigint,
    in_origin_sha bytea,
    in_require_phishing_resistant boolean
)
RETURNS TABLE (
    revision bigint,
    issued_at timestamptz,
    expires_at timestamptz,
    credential_id bytea,
    rp_id text,
    user_verification_required boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        ''::bytea,
        NULL::text,
        NULL::boolean
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauthentication_start_v1(name, uuid, bytea, uuid, uuid, uuid, bigint, bigint, text, uuid, text, uuid, text, bytea, bytea, bytea, bigint, bytea, boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_reauthentication_start_v1(name, uuid, bytea, uuid, uuid, uuid, bigint, bigint, text, uuid, text, uuid, text, bytea, bytea, bytea, bigint, bytea, boolean) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_reauthentication_start_v1(name, uuid, bytea, uuid, uuid, uuid, bigint, bigint, text, uuid, text, uuid, text, bytea, bytea, bytea, bigint, bytea, boolean);
