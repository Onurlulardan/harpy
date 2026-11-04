-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_worker_commit_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_lease_generation bigint,
    in_resource text,
    in_token bigint,
    in_authorization bigint,
    in_document bytea,
    in_source_bytes bigint
)
RETURNS TABLE (
    next_row_index bigint,
    processed_rows bigint,
    created_rows bigint,
    updated_rows bigint,
    skipped_rows bigint,
    rejected_rows bigint,
    decompressed_bytes bigint
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
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_worker_commit_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bytea, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_commit_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bytea, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_import_worker_commit_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bytea, bigint);
