-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_snapshot_accept_v1(
    p_operation_id uuid,
    p_lease_token uuid,
    p_license_id uuid,
    p_organization_id uuid,
    p_environment_id uuid,
    p_delivery_kind text,
    p_issuer text,
    p_key_id text,
    p_digest text,
    p_status text,
    p_sequence bigint,
    p_issued_at timestamptz,
    p_not_before timestamptz,
    p_refresh_after timestamptz,
    p_license_grant_deadline timestamptz,
    p_grace_until timestamptz,
    p_commercial_expires_at timestamptz,
    p_channel text,
    p_features jsonb,
    p_limits jsonb,
    p_canonical_payload bytea,
    p_signed_document bytea,
    p_projection jsonb
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    WITH head AS (
        SELECT snapshot.sequence, snapshot.payload_sha256
        FROM platform.billing_license_snapshots AS snapshot
        WHERE snapshot.license_id = p_license_id
          AND snapshot.environment_id = p_environment_id
        ORDER BY snapshot.sequence DESC
        LIMIT 1
    ), decision AS (
        SELECT CASE
            WHEN head.sequence IS NULL OR p_sequence > head.sequence THEN 'accepted'
            WHEN head.sequence = p_sequence AND head.payload_sha256 = p_digest THEN 'noop'
            WHEN head.sequence = p_sequence THEN 'sequence_conflict'
            ELSE 'sequence_rollback'
        END AS outcome
        FROM (SELECT 1) AS singleton
        LEFT JOIN head ON true
    ), stored AS (
        INSERT INTO platform.billing_license_snapshots (
            license_id, environment_id, sequence, payload_sha256, issuer, status,
            channel, features, limits, accepted_at
        )
        SELECT
            p_license_id, p_environment_id, p_sequence, p_digest, p_issuer, p_status,
            p_channel, p_features, p_limits, pg_catalog.clock_timestamp()
        FROM decision
        WHERE decision.outcome = 'accepted'
        ON CONFLICT (license_id, environment_id, sequence) DO NOTHING
        RETURNING sequence
    )
    SELECT decision.outcome FROM decision;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_snapshot_accept_v1(
    uuid, uuid, uuid, uuid, uuid, text, text, text, text, text, bigint,
    timestamptz, timestamptz, timestamptz, timestamptz, timestamptz, timestamptz,
    text, jsonb, jsonb, bytea, bytea, jsonb
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_snapshot_accept_v1(
    uuid, uuid, uuid, uuid, uuid, text, text, text, text, text, bigint,
    timestamptz, timestamptz, timestamptz, timestamptz, timestamptz, timestamptz,
    text, jsonb, jsonb, bytea, bytea, jsonb
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_snapshot_accept_v1(
    uuid, uuid, uuid, uuid, uuid, text, text, text, text, text, bigint,
    timestamptz, timestamptz, timestamptz, timestamptz, timestamptz, timestamptz,
    text, jsonb, jsonb, bytea, bytea, jsonb
);
