-- +goose Up
CREATE TABLE platform.billing_license_refresh_operations (
    operation_id uuid PRIMARY KEY,
    license_id uuid,
    licensee_organization_id uuid,
    environment_id uuid,
    delivery_kind text,
    lease_token uuid,
    lease_expires_at timestamptz,
    nonce bytea,
    head_sequence bigint,
    head_sha256 text,
    state text,
    database_now timestamptz,
    created_at timestamptz,
    idempotency_key text,
    tenant_id uuid,
    actor_id uuid,
    request_sha256 text,
    failure_code text,
    completed_at timestamptz,
    revision bigint
);

CREATE UNIQUE INDEX billing_license_refresh_operations_idempotency
    ON platform.billing_license_refresh_operations (environment_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE TABLE platform.billing_license_snapshots (
    license_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    sequence bigint NOT NULL,
    payload_sha256 text NOT NULL,
    issuer text,
    status text,
    channel text,
    features jsonb,
    limits jsonb,
    accepted_at timestamptz,
    PRIMARY KEY (license_id, environment_id, sequence)
);

REVOKE ALL ON TABLE platform.billing_license_refresh_operations FROM PUBLIC;
REVOKE ALL ON TABLE platform.billing_license_snapshots FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON TABLE platform.billing_license_refresh_operations TO harpy_api, harpy_worker;
GRANT SELECT, INSERT, UPDATE ON TABLE platform.billing_license_snapshots TO harpy_api, harpy_worker;

-- 32 bytes from two random UUIDs. Foundation rejects pgcrypto, so gen_random_bytes is unavailable.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_refresh_claim_v1(
    p_operation_id uuid,
    p_license_id uuid,
    p_licensee_organization_id uuid,
    p_environment_id uuid,
    p_delivery_kind text,
    p_components text,
    p_lease_seconds bigint
)
RETURNS TABLE (
    operation_id uuid,
    lease_token uuid,
    license_id uuid,
    licensee_organization_id uuid,
    environment_id uuid,
    delivery_kind text,
    nonce bytea,
    head_found boolean,
    head_sequence bigint,
    head_sha256 text,
    database_now timestamptz,
    lease_expires_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    WITH clock AS (
        SELECT pg_catalog.clock_timestamp() AS database_now
    ), issued AS (
        SELECT
            pg_catalog.gen_random_uuid() AS lease_token,
            pg_catalog.uuid_send(pg_catalog.gen_random_uuid())
                || pg_catalog.uuid_send(pg_catalog.gen_random_uuid()) AS nonce,
            clock.database_now,
            clock.database_now + (p_lease_seconds * pg_catalog.interval '1 second') AS lease_expires_at
        FROM clock
    ), upserted AS (
        INSERT INTO platform.billing_license_refresh_operations AS operation (
            operation_id, license_id, licensee_organization_id, environment_id,
            delivery_kind, lease_token, lease_expires_at, nonce, state,
            database_now, created_at, revision
        )
        SELECT
            p_operation_id, p_license_id, p_licensee_organization_id, p_environment_id,
            p_delivery_kind, issued.lease_token, issued.lease_expires_at, issued.nonce,
            'leased', issued.database_now, issued.database_now, 1
        FROM issued
        ON CONFLICT (operation_id) DO UPDATE SET
            license_id = EXCLUDED.license_id,
            licensee_organization_id = EXCLUDED.licensee_organization_id,
            environment_id = EXCLUDED.environment_id,
            delivery_kind = EXCLUDED.delivery_kind,
            lease_token = EXCLUDED.lease_token,
            lease_expires_at = EXCLUDED.lease_expires_at,
            nonce = EXCLUDED.nonce,
            state = 'leased',
            database_now = EXCLUDED.database_now,
            revision = operation.revision + 1
        RETURNING
            operation.operation_id, operation.lease_token, operation.license_id,
            operation.licensee_organization_id, operation.environment_id,
            operation.delivery_kind, operation.nonce, operation.database_now,
            operation.lease_expires_at
    )
    SELECT
        upserted.operation_id, upserted.lease_token, upserted.license_id,
        upserted.licensee_organization_id, upserted.environment_id,
        upserted.delivery_kind, upserted.nonce,
        head.sequence IS NOT NULL,
        COALESCE(head.sequence, 0),
        COALESCE(head.payload_sha256, pg_catalog.repeat('0', 64)),
        upserted.database_now, upserted.lease_expires_at
    FROM upserted
    LEFT JOIN LATERAL (
        SELECT snapshot.sequence, snapshot.payload_sha256
        FROM platform.billing_license_snapshots AS snapshot
        WHERE snapshot.license_id = upserted.license_id
          AND snapshot.environment_id = upserted.environment_id
        ORDER BY snapshot.sequence DESC
        LIMIT 1
    ) AS head ON true;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_refresh_claim_v1(
    uuid, uuid, uuid, uuid, text, text, bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_refresh_claim_v1(
    uuid, uuid, uuid, uuid, text, text, bigint
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_refresh_claim_v1(
    uuid, uuid, uuid, uuid, text, text, bigint
);
DROP TABLE platform.billing_license_snapshots;
DROP TABLE platform.billing_license_refresh_operations;
