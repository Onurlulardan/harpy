-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_restore_binding_v1(
    p_environment_id uuid
)
RETURNS TABLE (
    license_id uuid,
    licensee_organization_id uuid,
    environment_id uuid,
    delivery_kind text,
    issuer text,
    sequence bigint,
    payload_sha256 text,
    authority_revision bigint,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        snapshot.license_id,
        operation.licensee_organization_id,
        snapshot.environment_id,
        operation.delivery_kind,
        snapshot.issuer,
        snapshot.sequence,
        snapshot.payload_sha256,
        snapshot.sequence,
        pg_catalog.clock_timestamp()
    FROM platform.billing_license_snapshots AS snapshot
    JOIN platform.billing_license_refresh_operations AS operation
      ON operation.license_id = snapshot.license_id
     AND operation.environment_id = snapshot.environment_id
    WHERE snapshot.environment_id = p_environment_id
      AND snapshot.sequence = (
          SELECT pg_catalog.max(latest.sequence)
          FROM platform.billing_license_snapshots AS latest
          WHERE latest.environment_id = p_environment_id
      )
    ORDER BY operation.database_now DESC
    LIMIT 1;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_restore_binding_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_restore_binding_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_restore_binding_v1(uuid);
