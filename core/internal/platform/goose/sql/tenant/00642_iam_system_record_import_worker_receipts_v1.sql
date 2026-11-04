-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_worker_receipts_v1(
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
    in_indexes bigint[]
)
RETURNS TABLE (
    row_index bigint,
    outcome text,
    owner_kind text,
    owner_id uuid,
    owner_revision bigint,
    error_code text,
    source_sha256 bytea,
    result_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        ''::text,
        NULL::text,
        NULL::uuid,
        NULL::bigint,
        NULL::text,
        ''::bytea,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_worker_receipts_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint[]) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_receipts_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint[]) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_import_worker_receipts_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint[]);
