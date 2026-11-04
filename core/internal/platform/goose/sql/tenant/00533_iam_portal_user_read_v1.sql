-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_portal_user_read_v1(
    in_schema_name name,
    in_user_id uuid
)
RETURNS TABLE (
    id uuid,
    binding_id uuid,
    email text,
    status text,
    account_type text,
    subject_type text,
    subject_record_id uuid,
    revision bigint,
    setup_required boolean,
    created_at timestamptz,
    updated_at timestamptz
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
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        false,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_user_read_v1(name, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_portal_user_read_v1(name, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_portal_user_read_v1(name, uuid);
