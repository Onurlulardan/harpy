-- +goose Up
CREATE VIEW platform.billing_catalog_offering_snapshot_read_v1
WITH (security_barrier = true)
AS
SELECT source.offering_id AS id,
       source.offering_key,
       source.revision,
       source.lifecycle,
       source.id AS version_id,
       source.version,
       source.publication_state,
       source.is_public,
       source.available,
       source.name,
       source.description,
       source.sort_order,
       source.highlight_label,
       source.color,
       source.billing_interval,
       source.is_base,
       source.trial_days,
       source.grace_days,
       source.expected_price_count,
       source.expected_grant_count,
       source.effective_start,
       source.effective_end,
       source.created_at,
       source.updated_at,
       COALESCE((
           SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
               'id', price.id, 'cadence', price.cadence,
               'minor_units', price.minor_units,
               'currency', pg_catalog.btrim(price.currency::text),
               'currency_registry_version', price.currency_registry_version
           ) ORDER BY price.cadence, price.currency, price.id)
           FROM platform.billing_price_versions AS price
           WHERE price.offering_version_id = source.id
       ), '[]'::jsonb) AS prices,
       COALESCE((
           SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
               'key', definition.entitlement_key,
               'kind', grant_row.value_type,
               'boolean', grant_row.boolean_value,
               'integer', grant_row.integer_value,
               'unlimited', grant_row.unlimited
           ) ORDER BY definition.entitlement_key, definition.id)
           FROM platform.billing_offering_entitlements AS grant_row
           JOIN platform.billing_entitlement_definitions AS definition
             ON definition.authority_id = grant_row.authority_id
            AND definition.id = grant_row.entitlement_definition_id
           WHERE grant_row.offering_version_id = source.id
       ), '[]'::jsonb) AS entitlements
FROM platform.billing_offering_query_v2 AS source;

CREATE VIEW platform.billing_catalog_campaign_snapshot_read_v1
WITH (security_barrier = true)
AS
SELECT campaign.*,
       version_row.publication_state
FROM platform.billing_promotion_campaign_read_v2 AS campaign
JOIN platform.billing_campaign_versions AS version_row
  ON version_row.authority_id = campaign.authority_id
 AND version_row.id = campaign.version_id
WHERE campaign.authority_id =
      platform.billing_resolve_current_global_authority_v1('billing'::text);

GRANT SELECT ON platform.billing_catalog_offering_snapshot_read_v1
    TO forgecrud_runtime;
GRANT SELECT ON platform.billing_catalog_campaign_snapshot_read_v1
    TO forgecrud_runtime;
