-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_session_issue_v2(
    in_schema_name name,
    in_user_id uuid,
    in_expected_user_revision bigint,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_access_mode_id uuid,
    in_access_mode text,
    in_expected_access_mode_revision bigint,
    in_portal_id uuid,
    in_portal_revision bigint,
    in_session_id uuid,
    in_kind text,
    in_browser_family text,
    in_os_family text,
    in_ip_key_id text,
    in_ip_hmac bytea,
    in_family_id uuid,
    in_token_id uuid,
    in_selector bytea,
    in_hash_key_id text,
    in_secret_hmac bytea
)
RETURNS TABLE (
    session_revision bigint,
    family_revision bigint,
    token_revision bigint,
    issued_at timestamptz,
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
AS $fn$
    SELECT
        0::bigint,
        0::bigint,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        0::bigint,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_session_issue_v2(name, uuid, bigint, bigint, bigint, uuid, text, bigint, uuid, bigint, uuid, text, text, text, text, bytea, uuid, uuid, bytea, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_session_issue_v2(name, uuid, bigint, bigint, bigint, uuid, text, bigint, uuid, bigint, uuid, text, text, text, text, bytea, uuid, uuid, bytea, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_session_issue_v2(name, uuid, bigint, bigint, bigint, uuid, text, bigint, uuid, bigint, uuid, text, text, text, text, bytea, uuid, uuid, bytea, text, bytea);
