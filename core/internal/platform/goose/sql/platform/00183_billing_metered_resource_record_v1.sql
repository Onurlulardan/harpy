-- +goose Up
-- BELIRSIZ: mutation persistence is not determined by the caller.
CREATE FUNCTION platform.billing_metered_resource_record_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource text,
    p_mutation text,
    p_operation_id uuid,
    p_owner_fact_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_feature_enabled boolean,
    p_allowed boolean,
    p_gauge_delta bigint,
    p_current_value bigint,
    p_limit_value bigint,
    p_unlimited boolean,
    p_usage_decision_id uuid,
    p_usage_entry_id uuid,
    p_before_revision bigint,
    p_recorded_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.billing_metered_resource_record_v1(
    uuid, uuid, uuid, text, text, uuid,
    uuid, text, bytea, boolean, boolean, bigint,
    bigint, bigint, boolean, uuid, uuid, bigint,
    timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_record_v1(
    uuid, uuid, uuid, text, text, uuid,
    uuid, text, bytea, boolean, boolean, bigint,
    bigint, bigint, boolean, uuid, uuid, bigint,
    timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_metered_resource_record_v1(
    uuid, uuid, uuid, text, text, uuid,
    uuid, text, bytea, boolean, boolean, bigint,
    bigint, bigint, boolean, uuid, uuid, bigint,
    timestamptz
);
