-- +goose Up
-- BELIRSIZ: state, revision, and replay are not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_place_legal_hold_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_hold_id uuid,
    p_operation_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_reason text,
    p_ticket text,
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
AS $fn$
    SELECT
        p_hold_id,
        p_tenant_id,
        ''::text,
        p_actor_id,
        p_operation_id,
        p_reason,
        p_ticket,
        p_occurred_at,
        NULL::uuid,
        NULL::uuid,
        NULL::text,
        NULL::timestamptz,
        0::bigint,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_place_legal_hold_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_place_legal_hold_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_place_legal_hold_v1(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, timestamptz
);
