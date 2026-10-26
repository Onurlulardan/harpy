WITH selected AS (
    SELECT campaign.*
    FROM platform.billing_catalog_campaign_snapshot_read_v1 AS campaign
    WHERE campaign.lifecycle = 'active'
      AND campaign.publication_state = 'published'
      AND campaign.available
      AND campaign.starts_at <= $1::timestamptz
      AND (campaign.ends_at IS NULL OR $1::timestamptz < campaign.ends_at)
    ORDER BY campaign.created_at DESC, campaign.campaign_id
    LIMIT 1
)
SELECT selected.campaign_id, selected.legacy_id, selected.created_by,
       selected.campaign_key, selected.revision,
       selected.version_id, selected.version, selected.name,
       selected.description, selected.banner_text, selected.auto_apply,
       selected.available, selected.priority, selected.starts_at,
       selected.ends_at, selected.created_at, selected.updated_at,
       CASE WHEN discount.version_id IS NULL THEN NULL ELSE
           pg_catalog.jsonb_build_object(
               'id', discount.discount_id, 'legacy_id', discount.legacy_id,
               'created_by', discount.created_by, 'code', discount.code,
               'revision', discount.revision, 'version_id', discount.version_id,
               'version', discount.version, 'state', discount.lifecycle,
               'name', discount.name, 'description', discount.description,
               'available', discount.available, 'kind', discount.discount_kind,
               'basis_points', discount.basis_points,
               'fixed_minor_units', discount.fixed_minor_units,
               'currency', pg_catalog.btrim(discount.currency),
               'currency_registry_version', discount.currency_registry_version,
               'duration', discount.duration,
               'duration_cycles', discount.repeating_cycles,
               'priority', discount.priority, 'all_offerings', discount.all_offerings,
               'offering_version_ids', pg_catalog.to_jsonb(discount.offering_version_ids),
               'stable_offering_ids', pg_catalog.to_jsonb(discount.offering_ids),
               'legacy_offering_ids', pg_catalog.to_jsonb(discount.legacy_offering_ids),
               'used_count', discount.used_count,
               'campaign_ids', pg_catalog.to_jsonb(discount.campaign_ids),
               'legacy_campaign_ids', pg_catalog.to_jsonb(discount.legacy_campaign_ids),
               'valid_start', discount.valid_start, 'valid_end', discount.valid_end,
               'global_use_limit', discount.global_use_limit,
               'per_tenant_use_limit', discount.per_tenant_use_limit,
               'created_at', discount.created_at, 'updated_at', discount.updated_at
           )
       END
FROM selected
LEFT JOIN platform.billing_promotion_discount_read_v2 AS discount
  ON discount.authority_id = selected.authority_id
 AND selected.campaign_id = ANY(discount.campaign_ids)
 AND discount.lifecycle = 'active' AND discount.available
 AND discount.valid_start <= $1::timestamptz
 AND (discount.valid_end IS NULL OR $1::timestamptz < discount.valid_end)
ORDER BY discount.priority, discount.version_id NULLS LAST;
