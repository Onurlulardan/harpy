-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_load_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_export_id uuid,
    in_job_id uuid,
    in_resource text
)
RETURNS TABLE (
    actor_id uuid,
    token_version bigint,
    authorization_version bigint,
    catalog_version_id uuid,
    catalog_sha256 bytea,
    query_document bytea,
    query_sha256 bytea,
    format text,
    state text,
    attempt_generation bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::bytea,
        ''::bytea,
        ''::bytea,
        ''::text,
        ''::text,
        NULL::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_load_v1(name, uuid, uuid, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_load_v1(name, uuid, uuid, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_worker_load_v1(name, uuid, uuid, uuid, text);
