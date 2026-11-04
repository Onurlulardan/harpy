-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_directory_repair_page_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea
)
RETURNS TABLE (
    id uuid,
    user_id uuid,
    kind text,
    normalized bytea,
    normalized_sha bytea,
    state text,
    source_revision bigint
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
        ''::bytea,
        ''::bytea,
        ''::text,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_directory_repair_page_v1(name, uuid, uuid, uuid, bigint, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_directory_repair_page_v1(name, uuid, uuid, uuid, bigint, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_directory_repair_page_v1(name, uuid, uuid, uuid, bigint, text, bytea);
