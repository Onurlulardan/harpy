-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_session_list_v1(
    in_schema_name name,
    in_user_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_limit bigint,
    in_after_last_seen_at timestamptz,
    in_after_id uuid
)
RETURNS TABLE (
    id uuid,
    access_mode text,
    portal_binding_id uuid,
    device_kind text,
    browser_family text,
    os_family text,
    revision bigint,
    created_at timestamptz,
    last_seen_at timestamptz,
    absolute_expiry timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        NULL::uuid,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_session_list_v1(name, uuid, uuid, uuid, bigint, bigint, bigint, timestamptz, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_session_list_v1(name, uuid, uuid, uuid, bigint, bigint, bigint, timestamptz, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_session_list_v1(name, uuid, uuid, uuid, bigint, bigint, bigint, timestamptz, uuid);
