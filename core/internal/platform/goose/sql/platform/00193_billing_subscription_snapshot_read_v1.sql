-- +goose Up
-- BELIRSIZ: current subscription snapshot is not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_subscription_snapshot_read_v1(
    p_authority_id uuid,
    p_tenant_id uuid
)
RETURNS TABLE (
    candidate_count bigint,
    id uuid,
    legacy_id uuid,
    tenant_id uuid,
    legacy_plan_id uuid,
    provider_account_id uuid,
    offering_version_id uuid,
    price_version_id uuid,
    provider_key text,
    state text,
    period_start timestamptz,
    period_end timestamptz,
    trial_ends_at timestamptz,
    grace_ends_at timestamptz,
    cancel_at_period_end boolean,
    cancel_at timestamptz,
    trial_started_at timestamptz,
    cancelled_at timestamptz,
    cancellation_reason text,
    effective_minor_units bigint,
    effective_currency text,
    effective_registry_version bigint,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    price_matches_offering boolean,
    offering_matches boolean,
    offering_document jsonb
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        NULL::bigint,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::text,
        NULL::text,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::boolean,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::text,
        NULL::bigint,
        NULL::text,
        NULL::bigint,
        NULL::bigint,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::boolean,
        NULL::boolean,
        NULL::jsonb
    WHERE false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_subscription_snapshot_read_v1(
    uuid, uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_subscription_snapshot_read_v1(
    uuid, uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_subscription_snapshot_read_v1(
    uuid, uuid
);
