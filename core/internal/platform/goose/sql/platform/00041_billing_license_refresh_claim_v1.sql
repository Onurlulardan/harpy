-- +goose Up
-- BELIRSIZ: lease, nonce, and head are not determined by the caller.
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
AS $$
    SELECT
        p_operation_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        p_license_id,
        p_licensee_organization_id,
        p_environment_id,
        p_delivery_kind,
        ''::bytea,
        false,
        0::bigint,
        ''::text,
        clock.database_now,
        clock.database_now
    FROM (SELECT pg_catalog.clock_timestamp() AS database_now) AS clock;
$$;

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
