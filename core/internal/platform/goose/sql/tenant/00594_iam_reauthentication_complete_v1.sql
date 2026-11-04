-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_reauthentication_complete_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_binding bytea,
    in_challenge_id uuid,
    in_expected_challenge_revision bigint,
    in_expected_attempt_count smallint,
    in_user_id uuid,
    in_session_id uuid,
    in_expected_user_revision bigint,
    in_expected_session_revision bigint,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_access_mode_id uuid,
    in_access_mode text,
    in_expected_access_mode_revision bigint,
    in_portal_id uuid,
    in_portal_revision bigint,
    in_family_id uuid,
    in_token_id uuid,
    in_expected_family_revision bigint,
    in_expected_token_revision bigint,
    in_authenticator_id uuid,
    in_expected_authenticator_revision bigint,
    in_expected_subtype_revision bigint,
    in_expected_password_credential_version bigint,
    in_expected_password_credential_revision bigint,
    in_purpose bytea,
    in_context_sha bytea,
    in_method text,
    in_web_authn_sign_count bigint,
    in_web_authn_backup_eligible boolean,
    in_web_authn_backup_state boolean,
    in_totp_time_step bigint,
    in_successor_id uuid,
    in_selector bytea,
    in_successor_hash_key_id text,
    in_secret_hmac bytea
)
RETURNS TABLE (
    issued_at timestamptz,
    challenge_revision bigint,
    family_revision bigint,
    session_revision bigint,
    consumed_token_revision bigint,
    successor_token_revision bigint,
    locator_selector bytea,
    locator_expected_revision bigint,
    locator_revision bigint,
    locator_source_sha256 bytea,
    locator_state text,
    locator_projected_at timestamptz,
    locator_retain_until timestamptz,
    successor_selector bytea,
    successor_revision bigint,
    successor_source_sha256 bytea,
    successor_state text,
    successor_issued_at timestamptz,
    successor_projected_at timestamptz,
    successor_family_expires_at timestamptz,
    successor_retain_until timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        pg_catalog.clock_timestamp(),
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        ''::bytea,
        NULL::bigint,
        NULL::bigint,
        ''::bytea,
        NULL::text,
        NULL::timestamptz,
        NULL::timestamptz,
        ''::bytea,
        NULL::bigint,
        ''::bytea,
        NULL::text,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauthentication_complete_v1(name, uuid, bytea, uuid, bigint, smallint, uuid, uuid, bigint, bigint, bigint, bigint, uuid, text, bigint, uuid, bigint, uuid, uuid, bigint, bigint, uuid, bigint, bigint, bigint, bigint, bytea, bytea, text, bigint, boolean, boolean, bigint, uuid, bytea, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_reauthentication_complete_v1(name, uuid, bytea, uuid, bigint, smallint, uuid, uuid, bigint, bigint, bigint, bigint, uuid, text, bigint, uuid, bigint, uuid, uuid, bigint, bigint, uuid, bigint, bigint, bigint, bigint, bytea, bytea, text, bigint, boolean, boolean, bigint, uuid, bytea, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_reauthentication_complete_v1(name, uuid, bytea, uuid, bigint, smallint, uuid, uuid, bigint, bigint, bigint, bigint, uuid, text, bigint, uuid, bigint, uuid, uuid, bigint, bigint, uuid, bigint, bigint, bigint, bigint, bytea, bytea, text, bigint, boolean, boolean, bigint, uuid, bytea, text, bytea);
