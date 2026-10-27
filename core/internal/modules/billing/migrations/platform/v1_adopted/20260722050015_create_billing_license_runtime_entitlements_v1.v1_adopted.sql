-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_runtime_entitlements_v1(
    p_environment_id uuid
)
RETURNS TABLE (
    license_id uuid, environment_id uuid, snapshot_sequence bigint,
    payload_sha256 text, runtime_state text, entitlement_key text,
    value_type text, boolean_value boolean, integer_value bigint,
    unlimited boolean, release_channel text, database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    RETURN QUERY
    WITH authority AS (
        SELECT snapshot.snapshot_id, snapshot.license_id,
               snapshot.environment_id, snapshot.sequence,
               snapshot.payload_sha256,
               CASE
                   WHEN snapshot.status <> 'active'
                     OR v_now > snapshot.grace_until THEN 'build_off'
                   WHEN v_now > snapshot.grant_valid_until THEN 'grace'
                   ELSE 'valid'
               END AS state
        FROM platform.billing_license_heads AS head
        JOIN platform.billing_license_snapshots AS snapshot
          ON snapshot.snapshot_id = head.snapshot_id
        WHERE head.environment_id = p_environment_id
    )
    SELECT authority.license_id, authority.environment_id,
           authority.sequence, authority.payload_sha256, authority.state,
           projection.entitlement_key, projection.value_type,
           CASE
               WHEN authority.state = 'build_off'
                AND projection.value_type = 'boolean'
                AND projection.entitlement_key NOT IN (
                    'workflow.enabled', 'api.access'
                ) THEN false
               ELSE projection.boolean_value
           END,
           CASE
               WHEN authority.state = 'build_off'
                AND projection.value_type = 'integer' THEN 0
               ELSE projection.integer_value
           END,
           CASE
               WHEN authority.state = 'build_off'
                AND projection.value_type = 'integer' THEN false
               ELSE projection.unlimited
           END,
           snapshot.release_channel, v_now
    FROM authority
    JOIN platform.billing_license_snapshots AS snapshot
      ON snapshot.snapshot_id = authority.snapshot_id
    JOIN platform.billing_license_snapshot_projections AS projection
      ON projection.snapshot_id = authority.snapshot_id
    ORDER BY projection.entitlement_key;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_runtime_entitlements_v1(uuid)
FROM PUBLIC;
