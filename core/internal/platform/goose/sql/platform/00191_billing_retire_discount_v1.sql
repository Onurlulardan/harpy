-- +goose Up
-- BELIRSIZ: version, revision, lifecycle, and replay are not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_retire_discount_v1(
    p_authority_id uuid,
    p_operation_id uuid,
    p_discount_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_expected_revision bigint,
    p_reason text,
    p_occurred_at timestamptz
)
RETURNS TABLE (
    discount_id uuid,
    version_id uuid,
    revision bigint,
    version bigint,
    lifecycle text,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        p_discount_id,
        NULL::uuid,
        0::bigint,
        NULL::bigint,
        ''::text,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_retire_discount_v1(
    uuid, uuid, uuid, uuid, text, bytea,
    bigint, text, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_retire_discount_v1(
    uuid, uuid, uuid, uuid, text, bytea,
    bigint, text, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_retire_discount_v1(
    uuid, uuid, uuid, uuid, text, bytea,
    bigint, text, timestamptz
);
