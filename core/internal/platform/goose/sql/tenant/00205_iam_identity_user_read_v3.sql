-- +goose Up
CREATE FUNCTION iam_identity_user_read_v3(
    in_schema_name name,
    in_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    status text,
    token_version bigint,
    authorization_version bigint,
    revision bigint,
    identifier_id uuid,
    email_display text,
    email_normalized text,
    role_ids uuid[],
    organization_ids uuid[],
    created_at timestamptz,
    updated_at timestamptz,
    tombstoned_at timestamptz,
    email_verified boolean,
    access_modes text[]
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_id,
        ''::text,
        0::bigint,
        0::bigint,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        '{}'::uuid[],
        '{}'::uuid[],
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        false,
        '{}'::text[]
    ;
$$;

REVOKE ALL ON FUNCTION iam_identity_user_read_v3(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_user_read_v3(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_user_read_v3(name, uuid, uuid, uuid, bigint, bigint);
