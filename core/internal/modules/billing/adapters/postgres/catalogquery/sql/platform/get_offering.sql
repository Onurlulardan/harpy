WITH selected AS (
    SELECT DISTINCT ON (snapshot.id) snapshot.*
    FROM platform.billing_catalog_offering_snapshot_read_v1 AS snapshot
    WHERE snapshot.offering_key = $1::text
      AND snapshot.lifecycle = 'active'
      AND snapshot.publication_state = 'published'
      AND snapshot.available AND snapshot.is_public
      AND snapshot.effective_start <= $2::timestamptz
      AND (snapshot.effective_end IS NULL OR $2::timestamptz < snapshot.effective_end)
    ORDER BY snapshot.id, snapshot.version DESC, snapshot.version_id
)
SELECT selected.id, selected.offering_key, selected.revision,
       selected.version_id, selected.version, selected.available,
       selected.name, selected.description, selected.sort_order,
       selected.highlight_label, selected.color, selected.billing_interval,
       selected.is_base, selected.trial_days, selected.grace_days,
       selected.expected_price_count, selected.expected_grant_count,
       selected.effective_start, selected.effective_end,
       selected.created_at, selected.updated_at,
       selected.prices, selected.entitlements
FROM selected;
