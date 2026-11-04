-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_platform_license_refresh_request_v1(
    p_environment_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 text
)
RETURNS TABLE (
    operation_id uuid,
    replayed boolean,
    created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    WITH existing AS (
        SELECT operation.operation_id, operation.created_at
        FROM platform.billing_license_refresh_operations AS operation
        WHERE operation.environment_id = p_environment_id
          AND operation.idempotency_key = p_idempotency_key
    ), created AS (
        INSERT INTO platform.billing_license_refresh_operations (
            operation_id, environment_id, tenant_id, actor_id, idempotency_key,
            request_sha256, state, created_at, database_now, revision
        )
        SELECT pg_catalog.uuidv7(), p_environment_id, p_tenant_id, p_actor_id,
               p_idempotency_key, p_request_sha256, 'pending',
               pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp(), 1
        WHERE NOT EXISTS (SELECT 1 FROM existing)
        RETURNING operation_id, created_at
    )
    SELECT created.operation_id, false, created.created_at FROM created
    UNION ALL
    SELECT existing.operation_id, true, existing.created_at FROM existing;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_platform_license_refresh_request_v1(
    uuid, uuid, uuid, text, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_platform_license_refresh_request_v1(
    uuid, uuid, uuid, text, text
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_platform_license_refresh_request_v1(uuid, uuid, uuid, text, text);
