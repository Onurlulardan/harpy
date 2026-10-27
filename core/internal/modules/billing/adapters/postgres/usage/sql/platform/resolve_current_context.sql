SELECT *
FROM platform.billing_usage_context_snapshot_read_v1($1::uuid, $2::uuid)
AS snapshot(
    base_count bigint,
    base_offering_version_id text,
    subscription_count bigint,
    subscription_id text,
    subscription_offering_version_id text
);
