-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_refresh_target_acquire_v1()
RETURNS TABLE (
    target_found boolean,
    ready boolean,
    operation_id uuid,
    license_id uuid,
    licensee_organization_id uuid,
    environment_id uuid,
    delivery_kind text,
    target_revision bigint,
    retry_at timestamptz,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    WITH clock AS (
        SELECT pg_catalog.clock_timestamp() AS database_now
    ), candidate AS (
        SELECT operation.operation_id, operation.license_id,
               operation.licensee_organization_id, operation.environment_id,
               operation.delivery_kind, COALESCE(operation.revision, 1) AS revision
        FROM platform.billing_license_refresh_operations AS operation
        WHERE operation.state = 'retryable'
           OR (operation.state = 'leased' AND operation.lease_expires_at <= pg_catalog.clock_timestamp())
        ORDER BY operation.database_now
        LIMIT 1
    )
    SELECT
        candidate.operation_id IS NOT NULL,
        candidate.operation_id IS NOT NULL,
        candidate.operation_id,
        candidate.license_id,
        candidate.licensee_organization_id,
        candidate.environment_id,
        candidate.delivery_kind,
        candidate.revision,
        clock.database_now,
        clock.database_now
    FROM clock
    LEFT JOIN candidate ON true
    WHERE candidate.operation_id IS NOT NULL
    UNION ALL
    SELECT
        false, false,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text, 0, clock.database_now, clock.database_now
    FROM clock
    WHERE NOT EXISTS (SELECT 1 FROM candidate);
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_refresh_target_acquire_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_refresh_target_acquire_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_refresh_target_acquire_v1();
