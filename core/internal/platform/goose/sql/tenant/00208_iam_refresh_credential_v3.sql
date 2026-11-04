-- +goose Up
CREATE FUNCTION iam_refresh_credential_v3(
    in_schema_name name,
    in_source_token_id uuid,
    in_source_family_id uuid,
    in_selector bytea,
    in_source_revision bigint,
    in_binding bytea
)
RETURNS TABLE (
    token_id uuid,
    token_revision bigint,
    token_state text,
    family_id uuid,
    family_revision bigint,
    family_state text,
    session_id uuid,
    session_revision bigint,
    session_state text,
    user_id uuid,
    user_revision bigint,
    user_status text,
    identity_kind text,
    token_version bigint,
    authorization_version bigint,
    access_mode_id uuid,
    access_mode text,
    access_mode_state text,
    access_mode_revision bigint,
    portal_binding_id uuid,
    portal_binding_state text,
    portal_binding_revision bigint,
    hash_key_id text,
    secret_hmac bytea,
    sequence bigint,
    auth_time timestamptz,
    acr text,
    amr_password boolean,
    amr_webauthn boolean,
    amr_totp boolean,
    amr_recovery boolean,
    step_up_purpose_sha256 bytea,
    idle_expires_at timestamptz,
    absolute_expires_at timestamptz,
    retain_until timestamptz,
    platform_authorization_clock bigint,
    security_epoch bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        ''::text,
        0::bigint,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        NULL::uuid,
        NULL::text,
        NULL::bigint,
        ''::text,
        ''::bytea,
        0::bigint,
        pg_catalog.clock_timestamp(),
        ''::text,
        false,
        false,
        false,
        false,
        ''::bytea,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        0::bigint,
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION iam_refresh_credential_v3(name, uuid, uuid, bytea, bigint, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_refresh_credential_v3(name, uuid, uuid, bytea, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_refresh_credential_v3(name, uuid, uuid, bytea, bigint, bytea);
