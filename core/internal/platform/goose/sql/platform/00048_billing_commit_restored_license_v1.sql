-- +goose Up
-- BELIRSIZ: accept, revision, and replay are not determined by the caller.
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
AS $$
    SELECT
        p_operation_id,
        p_restore_receipt_id,
        p_license_binding_id,
        p_previous_sequence,
        p_previous_sha256,
        0::bigint,
        ''::text,
        0::bigint,
        false;
$$;

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
