-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_record_export_reconcile_park_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_export_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_lease_generation bigint,
    in_resource text,
    in_effect_id uuid,
    in_effect_receipt_id uuid,
    in_request_sha256 bytea,
    in_started_at timestamptz,
    in_failure_sha256 bytea,
    in_observed_at timestamptz,
    in_code text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION iam_system_record_export_reconcile_park_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text, uuid, uuid, bytea, timestamptz, bytea, timestamptz, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_park_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text, uuid, uuid, bytea, timestamptz, bytea, timestamptz, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_reconcile_park_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text, uuid, uuid, bytea, timestamptz, bytea, timestamptz, text);
