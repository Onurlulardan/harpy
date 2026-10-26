WITH authority AS (
    SELECT platform.billing_resolve_current_global_authority_v1(
        'billing'::text
    ) AS authority_id
), alias AS (
    SELECT mapping.legacy_id, mapping.canonical_id, mapping.record_kind
    FROM authority
    JOIN platform.billing_legacy_identity_alias_v1 AS mapping
      ON mapping.authority_id = authority.authority_id
    WHERE mapping.scope_kind = 'global'
      AND mapping.record_kind = $1::text
      AND (mapping.legacy_id = $2::uuid OR mapping.canonical_id = $2::uuid)
), canonical AS (
    SELECT DISTINCT offering.id, 'offering'::text AS record_kind
    FROM platform.billing_catalog_offering_snapshot_read_v1 AS offering
    WHERE offering.id = $2::uuid
    UNION ALL
    SELECT discount.discount_id, 'discount'::text
    FROM authority
    JOIN platform.billing_promotion_discount_read_v2 AS discount
      ON discount.authority_id = authority.authority_id
    WHERE discount.discount_id = $2::uuid
    UNION ALL
    SELECT campaign.campaign_id, 'campaign'::text
    FROM platform.billing_catalog_campaign_snapshot_read_v1 AS campaign
    WHERE campaign.campaign_id = $2::uuid
)
SELECT alias.legacy_id, alias.canonical_id, alias.record_kind
FROM alias
UNION ALL
SELECT canonical.id, canonical.id, canonical.record_kind
FROM canonical
WHERE canonical.record_kind = $1::text
  AND pg_catalog.uuid_extract_version(canonical.id) = 7
  AND NOT EXISTS (SELECT 1 FROM alias);
