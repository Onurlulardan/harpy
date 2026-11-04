-- +goose Up
CREATE FUNCTION platform.billing_license_refresh_failure_v1(
    p_operation_id uuid,
    p_lease_token uuid,
    p_kind text
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    UPDATE platform.billing_license_refresh_operations AS operation
    SET state = CASE
            WHEN p_kind IN ('sequence_conflict', 'sequence_rollback') THEN 'fenced'
            ELSE 'retryable'
        END,
        failure_code = p_kind,
        database_now = pg_catalog.clock_timestamp()
    WHERE operation.operation_id = p_operation_id
      AND operation.lease_token = p_lease_token
    RETURNING CASE
        WHEN operation.state = 'fenced' THEN 'fenced'
        ELSE 'pending'
    END;
$$;

REVOKE ALL ON FUNCTION platform.billing_license_refresh_failure_v1(uuid, uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_refresh_failure_v1(uuid, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_refresh_failure_v1(uuid, uuid, text);
