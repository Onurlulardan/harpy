-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_read_v1(
    in_schema_name name,
    in_actor_id uuid,
    in_group_id uuid
)
RETURNS TABLE (
    id uuid,
    requester_id uuid,
    mode text,
    state text,
    reviewer_id uuid,
    task_id uuid,
    revision bigint,
    expires_at timestamptz,
    terminal_at timestamptz,
    created_at timestamptz,
    updated_at timestamptz,
    requests_json bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        NULL::uuid,
        NULL::uuid,
        0::bigint,
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_read_v1(name, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_request_read_v1(name, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_request_read_v1(name, uuid, uuid);
