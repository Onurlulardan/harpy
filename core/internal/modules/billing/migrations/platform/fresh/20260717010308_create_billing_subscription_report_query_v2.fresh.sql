-- +goose Up
-- Migration version: 20260717010308 UTC
CREATE VIEW platform.billing_subscription_report_query_v2
WITH (security_barrier = true)
AS
WITH current_global AS (
    SELECT authority.authority_id, authority.scope_id
    FROM platform.billing_global_authorities AS authority
    JOIN platform.environment_identity AS identity
      ON identity.environment_id = authority.scope_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority.scope_kind
     AND group_row.scope_id = authority.scope_id
     AND group_row.group_id = authority.authority_group_id
     AND group_row.generation = authority.authority_group_generation
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = authority.scope_kind
     AND domain_row.scope_id = authority.scope_id
     AND domain_row.domain = authority.authority_domain
     AND domain_row.generation = authority.authority_domain_generation
    WHERE authority.scope_kind = 'global'
      AND authority.authority_group_id = 'billing'
      AND authority.authority_domain = 'billing'
      AND group_row.writer = 'v2' AND group_row.state = 'v2'
      AND domain_row.writer = 'v2' AND domain_row.state = 'v2'
), current_tenants AS (
    SELECT authority.authority_id, authority.tenant_id
    FROM platform.billing_tenant_authorities AS authority
    JOIN platform.tenants AS tenant ON tenant.id = authority.tenant_id
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority.scope_kind
     AND group_row.scope_id = authority.scope_id
     AND group_row.group_id = authority.authority_group_id
     AND group_row.generation = authority.authority_group_generation
    JOIN platform.domain_cutovers AS billing_row
      ON billing_row.scope_kind = authority.scope_kind
     AND billing_row.scope_id = authority.scope_id
     AND billing_row.domain = authority.billing_domain
     AND billing_row.generation = authority.billing_generation
    JOIN platform.domain_cutovers AS forms_row
      ON forms_row.scope_kind = authority.scope_kind
     AND forms_row.scope_id = authority.scope_id
     AND forms_row.domain = authority.forms_domain
     AND forms_row.generation = authority.forms_generation
    WHERE group_row.writer = 'v2' AND group_row.state = 'v2'
      AND billing_row.writer = 'v2' AND billing_row.state = 'v2'
      AND forms_row.writer = 'v2' AND forms_row.state = 'v2'
)
SELECT
    subscription.id,
    COALESCE(subscription_alias.legacy_id, subscription.id) AS legacy_id,
    subscription.tenant_id,
    COALESCE(offering_alias.legacy_id, root.id) AS legacy_plan_id,
    subscription.offering_version_id,
    subscription.price_version_id,
    provider.provider_key AS provider,
    subscription.state,
    subscription.period_start,
    subscription.period_end,
    subscription.trial_ends_at,
    COALESCE(wire.trial_started_at,
        CASE WHEN subscription.state = 'trialing'
             THEN subscription.period_start END) AS trial_started_at,
    subscription.grace_ends_at,
    subscription.cancel_at_period_end,
    subscription.revision AS version,
    effective.effective_minor_units AS effective_price_minor_units,
    effective.effective_currency AS effective_price_currency,
    effective.effective_registry_version AS effective_price_registry_version,
    subscription.created_at,
    CASE WHEN wire.source_updated_at IS NULL
              OR subscription.updated_at >= wire.source_updated_at
         THEN subscription.updated_at
         ELSE wire.source_updated_at END AS updated_at,
    wire.cancelled_at,
    wire.cancellation_reason,
    pg_catalog.jsonb_build_object(
        'id', root.id,
        'key', root.offering_key,
        'revision', root.revision,
        'version_id', version_row.id,
        'version', version_row.version,
        'state', root.lifecycle,
        'name', COALESCE(version_row.presentation_metadata ->> 'name', ''),
        'description', COALESCE(
            version_row.presentation_metadata ->> 'description', ''
        ),
        'sort_order', COALESCE(
            (version_row.presentation_metadata ->> 'sort_order')::integer, 0
        ),
        'highlight_label', COALESCE(
            version_row.presentation_metadata ->> 'highlight_label', ''
        ),
        'color', COALESCE(version_row.presentation_metadata ->> 'color', ''),
        'billing_interval', COALESCE(
            version_row.presentation_metadata ->> 'billing_interval', 'monthly'
        ),
        'public', version_row.is_public,
        'base', version_row.is_base,
        'trial_days', version_row.trial_days,
        'grace_days', version_row.grace_days,
        'effective_start', version_row.effective_start,
        'effective_end', version_row.effective_end,
        'prices', COALESCE((
            SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object(
                    'id', price.id,
                    'cadence', price.cadence,
                    'amount', pg_catalog.jsonb_build_object(
                        'minor_units', price.minor_units,
                        'currency', pg_catalog.btrim(price.currency::text),
                        'currency_registry_version',
                            price.currency_registry_version
                    ),
                    'version', 1
                )
                ORDER BY price.cadence, price.currency, price.id
            )
            FROM platform.billing_price_versions AS price
            WHERE price.authority_id = global_authority.authority_id
              AND price.offering_version_id = version_row.id
        ), '[]'::jsonb),
        'entitlements', COALESCE((
            SELECT pg_catalog.jsonb_agg(
                pg_catalog.jsonb_build_object(
                    'key', definition.entitlement_key,
                    'kind', grant_row.value_type,
                    'boolean', grant_row.boolean_value,
                    'integer', grant_row.integer_value,
                    'unlimited', grant_row.unlimited
                )
                ORDER BY definition.entitlement_key, definition.id
            )
            FROM platform.billing_offering_entitlements AS grant_row
            JOIN platform.billing_entitlement_definitions AS definition
              ON definition.authority_id = grant_row.authority_id
             AND definition.id = grant_row.entitlement_definition_id
            WHERE grant_row.authority_id = global_authority.authority_id
              AND grant_row.offering_version_id = version_row.id
        ), '[]'::jsonb),
        'created_at', root.created_at,
        'updated_at', version_row.published_at
    ) AS offering
FROM current_global AS global_authority
JOIN current_tenants AS authority ON true
JOIN platform.billing_subscriptions AS subscription
  ON subscription.authority_id = authority.authority_id
 AND subscription.tenant_id = authority.tenant_id
JOIN platform.billing_provider_accounts AS provider
  ON provider.id = subscription.provider_account_id
JOIN platform.billing_offering_versions AS version_row
  ON version_row.authority_id = global_authority.authority_id
 AND version_row.id = subscription.offering_version_id
JOIN platform.billing_offerings AS root
  ON root.authority_id = global_authority.authority_id
 AND root.id = version_row.offering_id
LEFT JOIN platform.billing_v1_subscription_wire_archive AS wire
  ON wire.authority_id = subscription.authority_id
 AND wire.tenant_id = subscription.tenant_id
 AND wire.subscription_id = subscription.id
LEFT JOIN platform.billing_legacy_identity_alias_v1 AS subscription_alias
  ON subscription_alias.authority_id = subscription.authority_id
 AND subscription_alias.scope_kind = 'tenant'
 AND subscription_alias.scope_id = subscription.tenant_id
 AND subscription_alias.record_kind = 'subscription'
 AND subscription_alias.canonical_id = subscription.id
LEFT JOIN platform.billing_legacy_identity_alias_v1 AS offering_alias
  ON offering_alias.authority_id = global_authority.authority_id
 AND offering_alias.scope_kind = 'global'
 AND offering_alias.scope_id = global_authority.scope_id
 AND offering_alias.record_kind = 'offering'
 AND offering_alias.canonical_id = root.id
JOIN LATERAL platform.billing_effective_subscription_price_v1(
    subscription.authority_id,
    subscription.tenant_id,
    subscription.id,
    CURRENT_TIMESTAMP
) AS effective ON true
WHERE SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'forgecrud_runtime';

GRANT SELECT ON platform.billing_subscription_report_query_v2
    TO forgecrud_runtime;
