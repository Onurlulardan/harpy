-- +goose Up
CREATE FUNCTION iam_maintenance_profile_backfill_apply_v3(
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
    in_actual_schema_id uuid,
    in_revision bigint,
    in_sha bytea,
    in_canonical bytea,
    in_digest_2 bytea
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_apply_v3(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, uuid, uuid, bigint, bytea, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_backfill_apply_v3(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, uuid, uuid, bigint, bytea, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_profile_backfill_apply_v3(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint, uuid, uuid, bigint, bytea, bytea, bytea);
