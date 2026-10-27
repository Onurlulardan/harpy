-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_subscription_snapshot_read_v1(
    p_authority_id uuid, p_tenant_id uuid
)
RETURNS SETOF record
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
RETURN QUERY
WITH runtime_guard AS MATERIALIZED (
    SELECT 1
    WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
      AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_owner'
      AND p_authority_id IS NOT DISTINCT FROM
          platform.billing_resolve_current_tenant_authority_v1(p_tenant_id)
), global_authority AS MATERIALIZED (
    SELECT authority.authority_id, authority.scope_id
    FROM platform.billing_global_authorities AS authority
    JOIN platform.cutover_groups AS group_row
      ON group_row.scope_kind = authority.scope_kind
     AND group_row.scope_id = authority.scope_id
     AND group_row.group_id = authority.authority_group_id
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind = authority.scope_kind
     AND domain_row.scope_id = authority.scope_id
     AND domain_row.domain = authority.authority_domain
    WHERE authority.scope_kind = 'global'
      AND authority.scope_id = (
          SELECT identity.environment_id FROM platform.environment_identity AS identity
      )
      AND authority.authority_group_id = 'billing'
      AND authority.authority_domain = 'billing'
      AND authority.authority_group_generation = group_row.generation
      AND authority.authority_domain_generation = domain_row.generation
      AND group_row.writer = 'v2' AND group_row.state = 'v2'
      AND domain_row.writer = 'v2' AND domain_row.state = 'v2'
), candidates AS MATERIALIZED (
    SELECT subscription.*, pg_catalog.count(*) OVER () AS candidate_count
    FROM runtime_guard
    CROSS JOIN platform.billing_subscriptions AS subscription
    WHERE subscription.authority_id = p_authority_id
      AND subscription.tenant_id = p_tenant_id
      AND subscription.state IN ('trialing', 'active', 'past_due')
), chosen AS (
    SELECT * FROM candidates ORDER BY updated_at DESC, id ASC LIMIT 1
)
SELECT chosen.candidate_count, chosen.id,
       COALESCE(subscription_alias.legacy_id, chosen.id),
       chosen.tenant_id,
       COALESCE(offering_alias.legacy_id, root.id),
       chosen.provider_account_id, chosen.offering_version_id,
       chosen.price_version_id, provider.provider_key, chosen.state,
       chosen.period_start, chosen.period_end, chosen.trial_ends_at,
       chosen.grace_ends_at, chosen.cancel_at_period_end, chosen.cancel_at,
       COALESCE(wire.trial_started_at,
           CASE WHEN chosen.state = 'trialing' THEN chosen.period_start END),
       wire.cancelled_at, wire.cancellation_reason,
       effective.effective_minor_units, effective.effective_currency,
       effective.effective_registry_version,
       chosen.revision, chosen.created_at,
       CASE WHEN wire.source_updated_at IS NULL
                 OR chosen.updated_at >= wire.source_updated_at
            THEN chosen.updated_at ELSE wire.source_updated_at END,
       price.id IS NOT NULL,
       root.id IS NOT NULL AND version_row.id IS NOT NULL
         AND version_row.publication_state = 'published',
       CASE WHEN root.id IS NULL OR version_row.id IS NULL THEN NULL ELSE
         pg_catalog.jsonb_build_object(
           'ID', root.id, 'Key', root.offering_key, 'Revision', root.revision,
           'VersionID', version_row.id, 'Version', version_row.version,
           'State', root.lifecycle, 'Public', version_row.is_public,
           'Name', COALESCE(version_row.presentation_metadata ->> 'name', ''),
           'Description', COALESCE(version_row.presentation_metadata ->> 'description', ''),
           'SortOrder', COALESCE((version_row.presentation_metadata ->> 'sort_order')::integer, 0),
           'HighlightLabel', COALESCE(version_row.presentation_metadata ->> 'highlight_label', ''),
           'Color', COALESCE(version_row.presentation_metadata ->> 'color', ''),
           'BillingInterval', COALESCE(
               version_row.presentation_metadata ->> 'billing_interval', 'monthly'),
           'Base', version_row.is_base, 'TrialDays', version_row.trial_days,
           'GraceDays', version_row.grace_days,
           'ExpectedPriceCount', version_row.expected_price_count,
           'ExpectedGrantCount', version_row.expected_grant_count,
           'EffectiveStart', version_row.effective_start,
           'EffectiveEnd', version_row.effective_end,
           'CreatedAt', root.created_at, 'UpdatedAt', version_row.published_at,
           'Prices', COALESCE((
             SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
               'ID', child.id, 'Cadence', child.cadence,
               'MinorUnits', child.minor_units,
               'Currency', pg_catalog.btrim(child.currency::text),
               'CurrencyRegistryVersion', child.currency_registry_version
             ) ORDER BY child.cadence, child.currency, child.id)
             FROM platform.billing_price_versions AS child
             WHERE child.authority_id = global_authority.authority_id
               AND child.offering_version_id = version_row.id
           ), '[]'::jsonb),
           'Entitlements', COALESCE((
             SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
               'Key', definition.entitlement_key, 'Kind', grant_row.value_type,
               'Boolean', grant_row.boolean_value, 'Integer', grant_row.integer_value,
               'Unlimited', grant_row.unlimited
             ) ORDER BY definition.entitlement_key, definition.id)
             FROM platform.billing_offering_entitlements AS grant_row
             JOIN platform.billing_entitlement_definitions AS definition
               ON definition.authority_id = grant_row.authority_id
              AND definition.id = grant_row.entitlement_definition_id
             WHERE grant_row.authority_id = global_authority.authority_id
               AND grant_row.offering_version_id = version_row.id
           ), '[]'::jsonb)
         ) END
FROM chosen
JOIN platform.billing_effective_subscription_price_snapshot_read_v1(
  chosen.authority_id, chosen.tenant_id, chosen.id, CURRENT_TIMESTAMP
) AS effective ON true
LEFT JOIN global_authority ON true
LEFT JOIN platform.billing_offering_versions AS version_row
  ON version_row.authority_id = global_authority.authority_id
 AND version_row.id = chosen.offering_version_id
LEFT JOIN platform.billing_offerings AS root
  ON root.authority_id = global_authority.authority_id
 AND root.id = version_row.offering_id
LEFT JOIN platform.billing_v1_subscription_wire_archive AS wire
  ON wire.authority_id = chosen.authority_id
 AND wire.tenant_id = chosen.tenant_id AND wire.subscription_id = chosen.id
LEFT JOIN platform.billing_legacy_identity_alias_v1 AS subscription_alias
  ON subscription_alias.authority_id = chosen.authority_id
 AND subscription_alias.scope_kind = 'tenant'
 AND subscription_alias.scope_id = chosen.tenant_id
 AND subscription_alias.record_kind = 'subscription'
 AND subscription_alias.canonical_id = chosen.id
LEFT JOIN platform.billing_legacy_identity_alias_v1 AS offering_alias
  ON offering_alias.authority_id = global_authority.authority_id
 AND offering_alias.scope_kind = 'global'
 AND offering_alias.scope_id = global_authority.scope_id
 AND offering_alias.record_kind = 'offering'
 AND offering_alias.canonical_id = root.id
LEFT JOIN platform.billing_provider_accounts AS provider
  ON provider.id = chosen.provider_account_id
LEFT JOIN platform.billing_price_versions AS price
  ON price.authority_id = global_authority.authority_id
 AND price.id = chosen.price_version_id
 AND price.offering_version_id = chosen.offering_version_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.billing_subscription_snapshot_read_v1(uuid,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.billing_subscription_snapshot_read_v1(uuid,uuid)
    TO harpy_runtime;
