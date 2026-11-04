-- +goose Up
-- BELIRSIZ: decision, actor, provider, active job, state, revision, reference, and error are not determined by the caller.
CREATE FUNCTION platform.billing_finalize_provider_command_v1(
    p_tenant_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_lease_generation bigint,
    p_payload_revision integer,
    p_payload_reference text,
    p_payload_sha256 text,
    p_command_id uuid,
    p_next_finalizer_job_id uuid,
    p_reconciliation_job_id uuid,
    p_observed_at timestamptz
)
RETURNS TABLE (
    decision text,
    command_id uuid,
    actor_id uuid,
    provider_key text,
    active_job_id uuid,
    finalizer_job_id uuid,
    request_sha256 text,
    state text,
    revision bigint,
    occurred_at timestamptz,
    provider_reference text,
    error_code text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        p_command_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        p_next_finalizer_job_id,
        p_payload_sha256,
        ''::text,
        0::bigint,
        p_observed_at,
        NULL::text,
        NULL::text;
$$;

REVOKE ALL ON FUNCTION platform.billing_finalize_provider_command_v1(
    uuid, uuid, uuid, bigint, integer, text,
    text, uuid, uuid, uuid, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_finalize_provider_command_v1(
    uuid, uuid, uuid, bigint, integer, text,
    text, uuid, uuid, uuid, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_finalize_provider_command_v1(
    uuid, uuid, uuid, bigint, integer, text,
    text, uuid, uuid, uuid, timestamptz
);
