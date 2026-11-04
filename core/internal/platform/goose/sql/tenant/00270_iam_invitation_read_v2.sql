-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_invitation_read_v2(
    in_schema_name name,
    in_actor_id uuid,
    in_id uuid
)
RETURNS TABLE (
    id uuid,
    display_value text,
    normalized_value text,
    role_id uuid,
    role_name text,
    scope_kind text,
    id_2 uuid,
    organization_id uuid,
    invited_by uuid,
    accepted_user_id uuid,
    accepted_identifier_id uuid,
    state text,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    expires_at timestamptz,
    terminal_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_id,
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        NULL::uuid,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_invitation_read_v2(name, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_read_v2(name, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_read_v2(name, uuid, uuid);
