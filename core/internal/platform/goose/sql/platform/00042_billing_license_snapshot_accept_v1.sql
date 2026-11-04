-- +goose Up
-- BELIRSIZ: accepted, noop, rollback, and conflict are not determined by the caller.
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
AS $$
    SELECT ''::text;
$$;

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
