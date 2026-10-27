SELECT snapshot.id, snapshot.offering_key, snapshot.revision,
       snapshot.version_id, snapshot.version, snapshot.lifecycle,
       snapshot.is_public, snapshot.available,
       snapshot.name, snapshot.description, snapshot.sort_order,
       snapshot.highlight_label, snapshot.color, snapshot.billing_interval,
       snapshot.is_base, snapshot.trial_days, snapshot.grace_days,
       snapshot.expected_price_count, snapshot.expected_grant_count,
       snapshot.effective_start, snapshot.effective_end,
       snapshot.created_at, snapshot.updated_at,
       snapshot.prices, snapshot.entitlements
FROM platform.billing_catalog_offering_snapshot_read_v1 AS snapshot
WHERE snapshot.id = $1::uuid
ORDER BY snapshot.version DESC, snapshot.version_id
LIMIT 1;
