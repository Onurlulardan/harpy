-- +goose Up
CREATE FUNCTION iam_maintenance_identifier_retire_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea,
    in_next_id uuid
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

REVOKE ALL ON FUNCTION iam_maintenance_identifier_retire_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_identifier_retire_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_identifier_retire_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid);
