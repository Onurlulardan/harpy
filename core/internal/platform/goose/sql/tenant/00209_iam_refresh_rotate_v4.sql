-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_rotate_v4(
    in_schema_name name,
    in_token_id uuid,
    in_family_id uuid,
    in_expected_token_revision bigint,
    in_expected_family_revision bigint,
    in_session_id uuid,
    in_expected_session_revision bigint,
    in_user_id uuid,
    in_expected_user_revision bigint,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_access_mode_id uuid,
    in_expected_access_mode_revision bigint,
    in_portal_id uuid,
    in_portal_revision bigint,
    in_successor_id uuid,
    in_selector bytea,
    in_successor_hash_key_id text,
    in_secret_hmac bytea,
    in_expected_platform_authorization_clock bigint,
    in_expected_security_epoch bytea
)
RETURNS TABLE (
    outcome text,
    issued_at timestamptz,
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
        ''::text,
        NULL::timestamptz,
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

REVOKE ALL ON FUNCTION iam_refresh_rotate_v4(name, uuid, uuid, bigint, bigint, uuid, bigint, uuid, bigint, bigint, bigint, uuid, bigint, uuid, bigint, uuid, bytea, text, bytea, bigint, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_refresh_rotate_v4(name, uuid, uuid, bigint, bigint, uuid, bigint, uuid, bigint, bigint, bigint, uuid, bigint, uuid, bigint, uuid, bytea, text, bytea, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_refresh_rotate_v4(name, uuid, uuid, bigint, bigint, uuid, bigint, uuid, bigint, bigint, bigint, uuid, bigint, uuid, bigint, uuid, bytea, text, bytea, bigint, bytea);
