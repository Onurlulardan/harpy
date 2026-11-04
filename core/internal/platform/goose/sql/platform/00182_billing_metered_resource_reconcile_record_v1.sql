-- +goose Up
-- BELIRSIZ: reconciliation persistence is not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_metered_resource_reconcile_record_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource text,
    p_operation_id uuid,
    p_owner_fact_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_evidence_sha256 bytea,
    p_exists boolean,
    p_active boolean,
    p_included_in_opening boolean,
    p_changed boolean,
    p_gauge_delta bigint,
    p_current_value bigint,
    p_limit_value bigint,
    p_unlimited boolean,
    p_usage_decision_id uuid,
    p_usage_entry_id uuid,
    p_before_revision bigint,
    p_recorded_at timestamptz,
    p_measured_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_metered_resource_reconcile_record_v1(
    uuid, uuid, uuid, text, uuid, uuid,
    text, bytea, bytea, boolean, boolean, boolean,
    boolean, bigint, bigint, bigint, boolean, uuid,
    uuid, bigint, timestamptz, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_reconcile_record_v1(
    uuid, uuid, uuid, text, uuid, uuid,
    text, bytea, bytea, boolean, boolean, boolean,
    boolean, bigint, bigint, bigint, boolean, uuid,
    uuid, bigint, timestamptz, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_metered_resource_reconcile_record_v1(
    uuid, uuid, uuid, text, uuid, uuid,
    text, bytea, bytea, boolean, boolean, boolean,
    boolean, bigint, bigint, bigint, boolean, uuid,
    uuid, bigint, timestamptz, timestamptz
);
