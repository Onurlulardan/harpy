SELECT *
FROM platform.billing_usage_snapshot_read_v1(
    $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid
)
AS snapshot(
    period_count bigint,
    override_count bigint,
    period_id uuid,
    entitlement_key text,
    value_type text,
    meter_kind text,
    reset_policy text,
    enforcement text,
    unlimited_allowed boolean,
    period_start timestamptz,
    period_end timestamptz,
    period_revision bigint,
    committed_total bigint,
    reserved_total bigint,
    gauge_total bigint,
    total_revision bigint,
    updated_at timestamptz,
    limit_amount bigint,
    unlimited boolean,
    limit_present boolean
);
