-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_platform_license_refresh_read_v1(
    p_environment_id uuid,
    p_operation_id uuid
)
RETURNS TABLE (
    operation_id uuid,
    operation_state text,
    snapshot_sequence bigint,
    snapshot_sha256 text,
    failure_code text,
    created_at timestamptz,
    completed_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT operation.operation_id, operation.state, operation.head_sequence,
           operation.head_sha256, operation.failure_code, operation.created_at,
           operation.completed_at
    FROM platform.billing_license_refresh_operations AS operation
    WHERE operation.environment_id = p_environment_id
      AND operation.operation_id = p_operation_id;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_platform_license_refresh_read_v1(uuid, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_platform_license_refresh_read_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_platform_license_refresh_read_v1(uuid, uuid);
