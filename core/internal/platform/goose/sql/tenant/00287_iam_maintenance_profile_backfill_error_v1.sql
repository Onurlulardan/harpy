-- +goose Up
CREATE FUNCTION iam_maintenance_profile_backfill_error_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea,
    in_plan_id uuid,
    in_page_number bigint,
    in_id uuid,
    in_code text,
    in_evidence bytea
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_error_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, uuid, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_error_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, uuid, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_profile_backfill_error_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, uuid, text, bytea);
