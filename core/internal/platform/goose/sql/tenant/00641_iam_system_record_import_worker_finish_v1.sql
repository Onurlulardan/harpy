-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_record_import_worker_finish_v1(
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
    in_success boolean
)
RETURNS TABLE (
    state text,
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
AS $$
    SELECT
        ''::text,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_system_record_import_worker_finish_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_finish_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, boolean) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_import_worker_finish_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, boolean);
