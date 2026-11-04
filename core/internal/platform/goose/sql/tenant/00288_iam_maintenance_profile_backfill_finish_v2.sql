-- +goose Up
CREATE FUNCTION iam_maintenance_profile_backfill_finish_v2(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea,
    in_plan_id uuid,
    in_page_number bigint,
    in_scanned bigint,
    in_converted bigint,
    in_source_sha bytea,
    in_result_sha bytea
)
RETURNS TABLE (
    job_id uuid,
    run_at timestamptz,
    reference text,
    digest bytea,
    dedupe text,
    priority integer
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_job_id,
        NULL::timestamptz,
        NULL::text,
        in_digest,
        NULL::text,
        NULL::integer
    ;
$$;

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_finish_v2(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, bigint, bigint, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_finish_v2(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, bigint, bigint, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_profile_backfill_finish_v2(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, bigint, bigint, bytea, bytea);
