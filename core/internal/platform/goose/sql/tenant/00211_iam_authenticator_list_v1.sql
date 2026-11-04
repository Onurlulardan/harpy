-- +goose Up
CREATE FUNCTION iam_authenticator_list_v1(
    in_schema_name name,
    in_user_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_limit bigint,
    in_after_updated_at timestamptz,
    in_after_id uuid
)
RETURNS TABLE (
    id uuid,
    user_id uuid,
    kind text,
    display_label text,
    state text,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    activated_at timestamptz,
    last_used_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_user_id,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        NULL::timestamptz
    ;
$$;

REVOKE ALL ON FUNCTION iam_authenticator_list_v1(name, uuid, uuid, uuid, bigint, bigint, bigint, timestamptz, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_authenticator_list_v1(name, uuid, uuid, uuid, bigint, bigint, bigint, timestamptz, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_authenticator_list_v1(name, uuid, uuid, uuid, bigint, bigint, bigint, timestamptz, uuid);
