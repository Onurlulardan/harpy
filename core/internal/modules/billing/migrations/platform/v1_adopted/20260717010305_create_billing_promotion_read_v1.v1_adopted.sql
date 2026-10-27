-- +goose Up
-- Migration version: 20260717010305 UTC
CREATE VIEW platform.billing_promotion_discount_read_v1
WITH (security_barrier = true)
AS
SELECT root.authority_id,
       root.id AS discount_id,
       root.normalized_code AS code,
       root.revision,
       root.lifecycle,
       root.created_at,
       GREATEST(root.created_at, version_row.published_at,
                COALESCE(root.revoked_at, root.created_at)) AS updated_at,
       version_row.id AS version_id,
       version_row.version,
       version_row.presentation_name AS name,
       version_row.presentation_description AS description,
       version_row.available,
       version_row.discount_kind,
       version_row.fixed_minor_units,
       version_row.currency::text AS currency,
       version_row.currency_registry_version,
       version_row.basis_points,
       version_row.duration,
       version_row.repeating_cycles,
       version_row.priority,
       version_row.all_offerings,
       version_row.valid_start,
       version_row.valid_end,
       version_row.global_use_limit,
       version_row.per_tenant_use_limit,
       COALESCE((
           SELECT pg_catalog.count(*)
           FROM platform.billing_discount_redemptions AS redemption
           JOIN platform.billing_discount_versions AS redeemed_version
             ON redeemed_version.id = redemption.discount_version_id
           WHERE redeemed_version.discount_id = root.id
       ), 0) AS used_count,
       COALESCE((
           SELECT pg_catalog.array_agg(
               binding.offering_version_id ORDER BY binding.offering_version_id
           )
           FROM platform.billing_discount_offerings AS binding
           WHERE binding.authority_id = root.authority_id
             AND binding.discount_version_id = version_row.id
       ), ARRAY[]::uuid[]) AS offering_version_ids,
       COALESCE((
           SELECT pg_catalog.array_agg(
               offering.id ORDER BY offering.id
           )
           FROM platform.billing_discount_offerings AS binding
           JOIN platform.billing_offering_versions AS offering_version
             ON offering_version.authority_id = binding.authority_id
            AND offering_version.id = binding.offering_version_id
           JOIN platform.billing_offerings AS offering
             ON offering.authority_id = offering_version.authority_id
            AND offering.id = offering_version.offering_id
           WHERE binding.authority_id = root.authority_id
             AND binding.discount_version_id = version_row.id
       ), ARRAY[]::uuid[]) AS offering_ids,
       COALESCE((
           SELECT pg_catalog.array_agg(
               campaign_version.campaign_id ORDER BY campaign_version.campaign_id
           )
           FROM platform.billing_campaign_discounts AS binding
           JOIN platform.billing_campaign_versions AS campaign_version
             ON campaign_version.authority_id = binding.authority_id
            AND campaign_version.id = binding.campaign_version_id
           WHERE binding.authority_id = root.authority_id
             AND binding.discount_version_id = version_row.id
             AND NOT EXISTS (
                 SELECT 1
                 FROM platform.billing_campaign_versions AS successor
                 WHERE successor.authority_id = campaign_version.authority_id
                   AND successor.superseded_version_id = campaign_version.id
             )
       ), ARRAY[]::uuid[]) AS campaign_ids
FROM platform.billing_discounts AS root
JOIN LATERAL (
    SELECT candidate.*
    FROM platform.billing_discount_versions AS candidate
    WHERE candidate.authority_id = root.authority_id
      AND candidate.discount_id = root.id
    ORDER BY candidate.version DESC, candidate.id ASC
    LIMIT 1
) AS version_row ON true
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';

GRANT SELECT ON platform.billing_promotion_discount_read_v1
TO harpy_runtime;

CREATE VIEW platform.billing_promotion_campaign_read_v1
WITH (security_barrier = true)
AS
SELECT root.authority_id,
       root.id AS campaign_id,
       root.campaign_key,
       root.revision,
       root.lifecycle,
       root.created_at,
       GREATEST(root.created_at, version_row.published_at,
                COALESCE(root.retired_at, root.created_at)) AS updated_at,
       version_row.id AS version_id,
       version_row.version,
       version_row.presentation_name AS name,
       version_row.presentation_description AS description,
       version_row.presentation_banner_text AS banner_text,
       version_row.auto_apply,
       version_row.available,
       version_row.priority,
       version_row.starts_at,
       version_row.ends_at
FROM platform.billing_campaigns AS root
JOIN LATERAL (
    SELECT candidate.*
    FROM platform.billing_campaign_versions AS candidate
    WHERE candidate.authority_id = root.authority_id
      AND candidate.campaign_id = root.id
      AND NOT EXISTS (
          SELECT 1
          FROM platform.billing_campaign_versions AS successor
          WHERE successor.authority_id = candidate.authority_id
            AND successor.superseded_version_id = candidate.id
      )
    ORDER BY candidate.version DESC, candidate.id ASC
    LIMIT 1
) AS version_row ON true
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';

GRANT SELECT ON platform.billing_promotion_campaign_read_v1
TO harpy_runtime;
