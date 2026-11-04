-- +goose Up
CREATE TABLE platform.billing_license_restore_receipts (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    restore_receipt_id uuid NOT NULL,
    restore_receipt_sha256 text,
    license_binding_id uuid NOT NULL,
    license_id uuid NOT NULL,
    licensee_organization_id uuid NOT NULL,
    previous_sequence bigint NOT NULL,
    previous_sha256 text NOT NULL,
    accepted_sequence bigint NOT NULL,
    accepted_sha256 text NOT NULL,
    authority_revision bigint NOT NULL,
    database_now timestamptz NOT NULL
);

REVOKE ALL ON TABLE platform.billing_license_restore_receipts FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON TABLE platform.billing_license_restore_receipts TO harpy_api, harpy_worker;

-- +goose StatementBegin
CREATE FUNCTION platform.billing_commit_restored_license_v1(
    p_operation_id uuid,
    p_restore_receipt_id uuid,
    p_restore_receipt_sha256 text,
    p_license_binding_id uuid,
    p_license_id uuid,
    p_licensee_organization_id uuid,
    p_environment_id uuid,
    p_previous_sequence bigint,
    p_previous_sha256 text,
    p_next_sequence bigint,
    p_next_sha256 text,
    p_expected_revision bigint,
    p_delivery_kind text,
    p_issuer text,
    p_key_id text,
    p_status text,
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
RETURNS TABLE (
    operation_id uuid,
    restore_receipt_id uuid,
    license_binding_id uuid,
    previous_sequence bigint,
    previous_sha256 text,
    accepted_sequence bigint,
    accepted_sha256 text,
    authority_revision bigint,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    WITH existing AS (
        SELECT receipt.operation_id, receipt.restore_receipt_id, receipt.license_binding_id,
               receipt.previous_sequence, receipt.previous_sha256, receipt.accepted_sequence,
               receipt.accepted_sha256, receipt.authority_revision
        FROM platform.billing_license_restore_receipts AS receipt
        WHERE receipt.operation_id = p_operation_id
          AND receipt.accepted_sha256 = p_next_sha256
          AND receipt.accepted_sequence = p_next_sequence
    ), stored_snapshot AS (
        INSERT INTO platform.billing_license_snapshots (
            license_id, environment_id, sequence, payload_sha256, issuer, status,
            channel, features, limits, accepted_at
        )
        SELECT p_license_id, p_environment_id, p_next_sequence, p_next_sha256, p_issuer,
               p_status, p_channel, p_features, p_limits, pg_catalog.clock_timestamp()
        WHERE NOT EXISTS (SELECT 1 FROM existing)
        ON CONFLICT (license_id, environment_id, sequence) DO NOTHING
        RETURNING sequence
    ), stored_receipt AS (
        INSERT INTO platform.billing_license_restore_receipts (
            operation_id, environment_id, restore_receipt_id, restore_receipt_sha256,
            license_binding_id, license_id, licensee_organization_id, previous_sequence,
            previous_sha256, accepted_sequence, accepted_sha256, authority_revision, database_now
        )
        SELECT
            p_operation_id, p_environment_id, p_restore_receipt_id, p_restore_receipt_sha256,
            p_license_binding_id, p_license_id, p_licensee_organization_id, p_previous_sequence,
            p_previous_sha256, p_next_sequence, p_next_sha256,
            CASE
                WHEN p_expected_revision > 1 THEN p_expected_revision
                ELSE 1::bigint
            END,
            pg_catalog.clock_timestamp()
        WHERE NOT EXISTS (SELECT 1 FROM existing)
        RETURNING operation_id, restore_receipt_id, license_binding_id, previous_sequence,
                  previous_sha256, accepted_sequence, accepted_sha256, authority_revision
    )
    SELECT stored_receipt.operation_id, stored_receipt.restore_receipt_id,
           stored_receipt.license_binding_id, stored_receipt.previous_sequence,
           stored_receipt.previous_sha256, stored_receipt.accepted_sequence,
           stored_receipt.accepted_sha256, stored_receipt.authority_revision, false
    FROM stored_receipt
    UNION ALL
    SELECT existing.operation_id, existing.restore_receipt_id, existing.license_binding_id,
           existing.previous_sequence, existing.previous_sha256, existing.accepted_sequence,
           existing.accepted_sha256, existing.authority_revision, true
    FROM existing;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_commit_restored_license_v1(
    uuid, uuid, text, uuid, uuid, uuid, uuid, bigint, text, bigint, text, bigint,
    text, text, text, text, timestamptz, timestamptz, timestamptz, timestamptz,
    timestamptz, timestamptz, text, jsonb, jsonb, bytea, bytea, jsonb
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_commit_restored_license_v1(
    uuid, uuid, text, uuid, uuid, uuid, uuid, bigint, text, bigint, text, bigint,
    text, text, text, text, timestamptz, timestamptz, timestamptz, timestamptz,
    timestamptz, timestamptz, text, jsonb, jsonb, bytea, bytea, jsonb
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_commit_restored_license_v1(
    uuid, uuid, text, uuid, uuid, uuid, uuid, bigint, text, bigint, text, bigint,
    text, text, text, text, timestamptz, timestamptz, timestamptz, timestamptz,
    timestamptz, timestamptz, text, jsonb, jsonb, bytea, bytea, jsonb
);
DROP TABLE platform.billing_license_restore_receipts;
