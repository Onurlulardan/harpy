-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_read_v1(
    in_schema_name name,
    in_transfer_id uuid,
    in_actor_id uuid,
    in_resource text,
    in_token_version bigint,
    in_authorization_version bigint
)
RETURNS TABLE (
    import_id uuid,
    job_id uuid,
    resource_kind text,
    state text,
    processed_rows bigint,
    created_rows bigint,
    updated_rows bigint,
    skipped_rows bigint,
    rejected_rows bigint,
    created_at timestamptz,
    completed_at timestamptz
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
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        pg_catalog.clock_timestamp(),
        NULL::timestamptz
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_read_v1(name, uuid, uuid, text, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_import_read_v1(name, uuid, uuid, text, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_import_read_v1(name, uuid, uuid, text, bigint, bigint);
