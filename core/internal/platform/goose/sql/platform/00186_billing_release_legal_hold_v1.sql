-- +goose Up
-- BELIRSIZ: placed identity, state, revision, and replay are not determined by the caller.
CREATE FUNCTION platform.billing_release_legal_hold_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_hold_id uuid,
    p_operation_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_reason text,
    p_occurred_at timestamptz
)
RETURNS TABLE (
    hold_id uuid,
    tenant_id uuid,
    state text,
    placed_by uuid,
    place_operation_id uuid,
    reason text,
    ticket text,
    placed_at timestamptz,
    released_by uuid,
    release_operation_id uuid,
    release_reason text,
    released_at timestamptz,
    revision bigint,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        p_hold_id,
        p_tenant_id,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        pg_catalog.clock_timestamp(),
        p_actor_id,
        p_operation_id,
        p_reason,
        p_occurred_at,
        0::bigint,
        false;
$$;

REVOKE ALL ON FUNCTION platform.billing_release_legal_hold_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_release_legal_hold_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_release_legal_hold_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, timestamptz
);
