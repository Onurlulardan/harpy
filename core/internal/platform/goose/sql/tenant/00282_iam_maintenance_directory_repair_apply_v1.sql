-- +goose Up
CREATE FUNCTION iam_maintenance_directory_repair_apply_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea,
    in_id uuid,
    in_user_id uuid,
    in_kind text,
    in_state text,
    in_source_revision bigint,
    in_normalized_sha bytea,
    in_evidence text
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION iam_maintenance_directory_repair_apply_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid, uuid, text, text, bigint, bytea, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_directory_repair_apply_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid, uuid, text, text, bigint, bytea, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_directory_repair_apply_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid, uuid, text, text, bigint, bytea, text);
