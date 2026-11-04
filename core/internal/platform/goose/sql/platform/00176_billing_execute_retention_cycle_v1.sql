-- +goose Up
-- BELIRSIZ: run time, dedupe key, priority, purge counts, and checkpoint are not determined by the caller.
CREATE FUNCTION platform.billing_execute_retention_cycle_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_next_job_id uuid,
    p_payload_reference text,
    p_payload_sha256 text
)
RETURNS TABLE (
    job_id uuid,
    run_at timestamptz,
    payload_reference text,
    payload_sha256 text,
    dedupe_key text,
    priority integer,
    raw_bodies_redacted bigint,
    raw_records_purged bigint,
    entitlement_purged bigint,
    financial_purged bigint,
    legal_hold_blocked bigint,
    checkpoint_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        p_next_job_id,
        pg_catalog.clock_timestamp(),
        p_payload_reference,
        p_payload_sha256,
        ''::text,
        0::integer,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.billing_execute_retention_cycle_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_execute_retention_cycle_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    text
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_execute_retention_cycle_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    text
);
