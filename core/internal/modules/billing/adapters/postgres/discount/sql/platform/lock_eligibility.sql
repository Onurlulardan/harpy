WITH candidate AS MATERIALIZED (
    SELECT version_row.id AS discount_version_id,
           discount_row.normalized_code,
           subscription.offering_version_id,
           subscription.price_version_id,
           price.currency::text AS required_currency,
           version_row.valid_start,
           COALESCE(
               version_row.valid_end,
               '9999-12-31 23:59:59.999999+00'::timestamptz
           ) AS valid_until,
           version_row.global_use_limit,
           version_row.per_tenant_use_limit,
           version_row.all_offerings,
           version_row.discount_kind,
           version_row.fixed_minor_units,
           version_row.currency,
           version_row.currency_registry_version,
           version_row.basis_points,
           version_row.duration,
           version_row.repeating_cycles,
           version_row.priority,
           version_row.version,
           version_row.available
    FROM platform.billing_tenant_authorities AS authority
    JOIN platform.billing_subscriptions AS subscription
      ON subscription.authority_id = authority.authority_id
     AND subscription.tenant_id = authority.tenant_id
    JOIN platform.billing_offering_versions AS offering
      ON offering.id = subscription.offering_version_id
    JOIN platform.billing_price_versions AS price
      ON price.id = subscription.price_version_id
     AND price.offering_version_id = offering.id
    JOIN platform.billing_discounts AS discount_row
      ON discount_row.normalized_code = $3
     AND discount_row.lifecycle = 'active'
    JOIN platform.billing_global_authorities AS global_authority
      ON global_authority.authority_id = discount_row.authority_id
     AND global_authority.scope_id = (
         SELECT environment_id FROM platform.environment_identity
     )
    JOIN platform.cutover_groups AS global_group
      ON global_group.scope_kind = global_authority.scope_kind
     AND global_group.scope_id = global_authority.scope_id
     AND global_group.group_id = global_authority.authority_group_id
     AND global_group.generation = global_authority.authority_group_generation
     AND global_group.state = 'v2' AND global_group.writer = 'v2'
    JOIN platform.domain_cutovers AS global_domain
      ON global_domain.scope_kind = global_authority.scope_kind
     AND global_domain.scope_id = global_authority.scope_id
     AND global_domain.domain = global_authority.authority_domain
     AND global_domain.generation = global_authority.authority_domain_generation
     AND global_domain.state = 'v2' AND global_domain.writer = 'v2'
    JOIN LATERAL (
        SELECT version_row.*
        FROM platform.billing_discount_versions AS version_row
        WHERE version_row.discount_id = discount_row.id
          AND version_row.valid_start <= $7
          AND (
              version_row.valid_end IS NULL
              OR version_row.valid_end > $7
          )
        ORDER BY version_row.version DESC
        LIMIT 1
    ) AS version_row ON true
    WHERE authority.authority_id = $1 AND authority.tenant_id = $2
      AND subscription.id = $4
      AND subscription.offering_version_id = $5
      AND subscription.price_version_id = $6
    FOR SHARE OF authority, subscription, offering, price, discount_row,
                 global_authority, global_group, global_domain, version_row
)
SELECT candidate.discount_version_id, candidate.normalized_code,
       candidate.offering_version_id, candidate.price_version_id,
       candidate.required_currency, candidate.valid_start, candidate.valid_until,
       candidate.global_use_limit,
       (SELECT pg_catalog.count(*)
        FROM platform.billing_discount_redemptions AS redemption
        WHERE redemption.discount_version_id = candidate.discount_version_id),
       candidate.per_tenant_use_limit,
       (SELECT pg_catalog.count(*)
        FROM platform.billing_discount_redemptions AS redemption
        WHERE redemption.discount_version_id = candidate.discount_version_id
          AND redemption.tenant_id = $2),
       (
           candidate.available
           AND (candidate.all_offerings OR EXISTS (
               SELECT 1 FROM platform.billing_discount_offerings AS allowed
               WHERE allowed.discount_version_id = candidate.discount_version_id
                 AND allowed.offering_version_id = candidate.offering_version_id
           ))
           AND (candidate.discount_kind <> 'fixed'
                OR candidate.currency::text = candidate.required_currency)
           AND NOT EXISTS (
               SELECT 1
               FROM platform.billing_discount_redemptions AS existing
               WHERE existing.subscription_id = $4
                 AND existing.discount_version_id = candidate.discount_version_id
           )
       ),
       pg_catalog.sha256(pg_catalog.convert_to(
           pg_catalog.jsonb_build_array(
               candidate.discount_version_id, candidate.version,
               candidate.discount_kind, candidate.fixed_minor_units,
               candidate.currency, candidate.currency_registry_version,
               candidate.basis_points, candidate.duration,
               candidate.repeating_cycles, candidate.priority,
               candidate.global_use_limit, candidate.per_tenant_use_limit,
               candidate.all_offerings, candidate.available
           )::text,
           'UTF8'
       ))
FROM candidate;
