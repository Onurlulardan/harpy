WITH locked AS MATERIALIZED (
    SELECT definition.entitlement_key, definition.value_type,
           definition.meter_kind, definition.reset_policy,
           definition.enforcement, definition.unlimited_allowed,
           CASE
             WHEN override_row.id IS NOT NULL THEN override_row.integer_value
             WHEN purchased.entitlement_definition_id IS NOT NULL
               THEN purchased.integer_value
             ELSE base_grant.integer_value
           END AS limit_amount,
           CASE
             WHEN override_row.id IS NOT NULL THEN override_row.unlimited
             WHEN purchased.entitlement_definition_id IS NOT NULL
               THEN purchased.unlimited
             ELSE base_grant.unlimited
           END AS unlimited,
           period.id AS period_id, period.period_start, period.period_end,
           total.committed_total, total.reserved_total, total.gauge_total,
           total.revision AS usage_revision
    FROM platform.billing_tenant_authorities AS authority
    JOIN platform.billing_meter_periods AS period
      ON period.authority_id = authority.authority_id
     AND period.tenant_id = authority.tenant_id
    JOIN platform.billing_usage_totals AS total
      ON total.authority_id = authority.authority_id
     AND total.tenant_id = authority.tenant_id
     AND total.meter_period_id = period.id
    JOIN platform.billing_entitlement_definitions AS definition
      ON definition.id = period.entitlement_definition_id
     AND definition.entitlement_key = period.entitlement_key
    JOIN LATERAL (
        SELECT assignment.offering_version_id
        FROM platform.billing_tenant_base_offerings AS assignment
        WHERE assignment.authority_id = authority.authority_id
          AND assignment.tenant_id = authority.tenant_id
          AND assignment.effective_start <= pg_catalog.statement_timestamp()
          AND (
              assignment.effective_end IS NULL
              OR assignment.effective_end > pg_catalog.statement_timestamp()
          )
        ORDER BY assignment.effective_start DESC
        LIMIT 1
    ) AS base_assignment ON true
    LEFT JOIN platform.billing_offering_entitlements AS base_grant
      ON base_grant.offering_version_id = base_assignment.offering_version_id
     AND base_grant.entitlement_definition_id = definition.id
    LEFT JOIN LATERAL (
        SELECT subscription.id, subscription.offering_version_id
        FROM platform.billing_subscriptions AS subscription
        WHERE subscription.authority_id = authority.authority_id
          AND subscription.tenant_id = authority.tenant_id
          AND (
              (subscription.state = 'trialing'
                AND pg_catalog.statement_timestamp() < subscription.trial_ends_at)
              OR (subscription.state = 'active'
                AND pg_catalog.statement_timestamp() < subscription.period_end)
              OR (subscription.state = 'past_due'
                AND pg_catalog.statement_timestamp() < subscription.grace_ends_at)
          )
        LIMIT 1
    ) AS subscription ON true
    LEFT JOIN platform.billing_offering_entitlements AS purchased
      ON purchased.offering_version_id = subscription.offering_version_id
     AND purchased.entitlement_definition_id = definition.id
    LEFT JOIN LATERAL (
        SELECT override_row.*
        FROM platform.billing_subscription_overrides AS override_row
        WHERE override_row.authority_id = authority.authority_id
          AND override_row.tenant_id = authority.tenant_id
          AND override_row.subscription_id = subscription.id
          AND override_row.entitlement_definition_id = definition.id
          AND override_row.effective_start <= pg_catalog.statement_timestamp()
          AND (
              override_row.effective_end IS NULL
              OR override_row.effective_end > pg_catalog.statement_timestamp()
          )
        ORDER BY override_row.effective_start DESC
        LIMIT 1
    ) AS override_row ON true
    WHERE authority.authority_id = $1 AND authority.tenant_id = $2
      AND period.entitlement_key = $3 AND period.id = $4
      AND period.state = 'open'
      AND (
          ($5 = 'reservation' AND period.meter_kind IN ('counter', 'gauge'))
          OR period.meter_kind = $5
      )
      AND definition.value_type = 'integer'
      AND (
          override_row.id IS NOT NULL
          OR purchased.entitlement_definition_id IS NOT NULL
          OR base_grant.entitlement_definition_id IS NOT NULL
      )
    FOR UPDATE OF total
)
SELECT entitlement_key, value_type, meter_kind, reset_policy, enforcement,
       unlimited_allowed, limit_amount, unlimited, period_id,
       period_start, period_end, committed_total, reserved_total, gauge_total,
       usage_revision
FROM locked;
