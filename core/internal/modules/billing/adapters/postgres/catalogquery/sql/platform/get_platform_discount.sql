SELECT pg_catalog.jsonb_build_object(
    'id', selected.discount_id, 'legacy_id', selected.legacy_id,
    'created_by', selected.created_by, 'code', selected.code,
    'revision', selected.revision, 'version_id', selected.version_id,
    'version', selected.version, 'state', selected.lifecycle,
    'name', selected.name, 'description', selected.description,
    'available', selected.available, 'kind', selected.discount_kind,
    'basis_points', selected.basis_points,
    'fixed_minor_units', selected.fixed_minor_units,
    'currency', pg_catalog.btrim(selected.currency),
    'currency_registry_version', selected.currency_registry_version,
    'duration', selected.duration, 'duration_cycles', selected.repeating_cycles,
    'priority', selected.priority, 'all_offerings', selected.all_offerings,
    'offering_version_ids', pg_catalog.to_jsonb(selected.offering_version_ids),
    'stable_offering_ids', pg_catalog.to_jsonb(selected.offering_ids),
    'legacy_offering_ids', pg_catalog.to_jsonb(selected.legacy_offering_ids),
    'used_count', selected.used_count,
    'campaign_ids', pg_catalog.to_jsonb(selected.campaign_ids),
    'legacy_campaign_ids', pg_catalog.to_jsonb(selected.legacy_campaign_ids),
    'valid_start', selected.valid_start, 'valid_end', selected.valid_end,
    'global_use_limit', selected.global_use_limit,
    'per_tenant_use_limit', selected.per_tenant_use_limit,
    'created_at', selected.created_at, 'updated_at', selected.updated_at
)
FROM platform.billing_promotion_discount_read_v2 AS selected
WHERE selected.authority_id =
      platform.billing_resolve_current_global_authority_v1('billing'::text)
  AND selected.discount_id = $1::uuid;
