-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_record_export_worker_checkpoint_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_export_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_attempt_generation bigint,
    in_resource text,
    in_token_version bigint,
    in_authorization_version bigint,
    in_rows bigint,
    in_plaintext_bytes bigint,
    in_checked_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION iam_system_record_export_worker_checkpoint_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint, bigint, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_checkpoint_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_worker_checkpoint_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint, bigint, timestamptz);
