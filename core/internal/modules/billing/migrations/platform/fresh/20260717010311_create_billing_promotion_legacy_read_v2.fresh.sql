-- +goose Up
-- Migration version: 20260717010311 UTC
ALTER TABLE platform.billing_discounts
    ADD COLUMN legacy_created_by uuid NOT NULL DEFAULT
        '00000000-0000-0000-0000-000000000000'::uuid;
ALTER TABLE platform.billing_campaigns
    ADD COLUMN legacy_created_by uuid NOT NULL DEFAULT
        '00000000-0000-0000-0000-000000000000'::uuid;

CREATE VIEW platform.billing_promotion_discount_read_v2
WITH (security_barrier = true)
AS
SELECT base.*,
       COALESCE(identity.legacy_id, base.discount_id) AS legacy_id,
       ARRAY(
           SELECT COALESCE(offering_identity.legacy_id, selected.offering_id)
           FROM pg_catalog.unnest(base.offering_ids) AS selected(offering_id)
           LEFT JOIN platform.billing_legacy_identity_alias_v1 AS offering_identity
             ON offering_identity.authority_id = base.authority_id
            AND offering_identity.record_kind = 'offering'
            AND offering_identity.canonical_id = selected.offering_id
           ORDER BY COALESCE(offering_identity.legacy_id, selected.offering_id)
       ) AS legacy_offering_ids,
       ARRAY(
           SELECT COALESCE(campaign_identity.legacy_id, selected.campaign_id)
           FROM pg_catalog.unnest(base.campaign_ids) AS selected(campaign_id)
           LEFT JOIN platform.billing_legacy_identity_alias_v1 AS campaign_identity
             ON campaign_identity.authority_id = base.authority_id
            AND campaign_identity.record_kind = 'campaign'
            AND campaign_identity.canonical_id = selected.campaign_id
           ORDER BY COALESCE(campaign_identity.legacy_id, selected.campaign_id)
       ) AS legacy_campaign_ids,
       root.legacy_created_by AS created_by
FROM platform.billing_promotion_discount_read_v1 AS base
JOIN platform.billing_discounts AS root
  ON root.authority_id = base.authority_id AND root.id = base.discount_id
LEFT JOIN platform.billing_legacy_identity_alias_v1 AS identity
  ON identity.authority_id = base.authority_id
 AND identity.record_kind = 'discount'
 AND identity.canonical_id = base.discount_id;

GRANT SELECT ON platform.billing_promotion_discount_read_v2
TO forgecrud_runtime;

CREATE VIEW platform.billing_promotion_campaign_read_v2
WITH (security_barrier = true)
AS
SELECT base.*,
       COALESCE(identity.legacy_id, base.campaign_id) AS legacy_id,
       root.legacy_created_by AS created_by
FROM platform.billing_promotion_campaign_read_v1 AS base
JOIN platform.billing_campaigns AS root
  ON root.authority_id = base.authority_id AND root.id = base.campaign_id
LEFT JOIN platform.billing_legacy_identity_alias_v1 AS identity
  ON identity.authority_id = base.authority_id
 AND identity.record_kind = 'campaign'
 AND identity.canonical_id = base.campaign_id;

GRANT SELECT ON platform.billing_promotion_campaign_read_v2
TO forgecrud_runtime;
