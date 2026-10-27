-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_usage_context_snapshot_read_v1(
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
), base_candidates AS MATERIALIZED (
    SELECT assignment.offering_version_id
    FROM runtime_guard
    CROSS JOIN platform.billing_tenant_base_offerings AS assignment
    WHERE assignment.authority_id = p_authority_id
      AND assignment.tenant_id = p_tenant_id
      AND assignment.effective_start <= pg_catalog.transaction_timestamp()
      AND (assignment.effective_end IS NULL
           OR assignment.effective_end > pg_catalog.transaction_timestamp())
), subscription_candidates AS MATERIALIZED (
    SELECT subscription.id, subscription.offering_version_id
    FROM runtime_guard
    CROSS JOIN platform.billing_subscriptions AS subscription
    WHERE subscription.authority_id = p_authority_id
      AND subscription.tenant_id = p_tenant_id
      AND subscription.period_start <= pg_catalog.transaction_timestamp()
      AND (
          (subscription.state = 'trialing'
            AND pg_catalog.transaction_timestamp() < subscription.trial_ends_at)
          OR (subscription.state = 'active'
            AND pg_catalog.transaction_timestamp() < subscription.period_end)
          OR (subscription.state = 'past_due'
            AND pg_catalog.transaction_timestamp() < subscription.grace_ends_at)
      )
)
SELECT
    (SELECT pg_catalog.count(*) FROM base_candidates),
    (SELECT offering_version_id::text FROM base_candidates
     ORDER BY offering_version_id LIMIT 1),
    (SELECT pg_catalog.count(*) FROM subscription_candidates),
    (SELECT id::text FROM subscription_candidates ORDER BY id LIMIT 1),
    (SELECT offering_version_id::text FROM subscription_candidates
     ORDER BY id LIMIT 1);
END;
$function$;

CREATE FUNCTION platform.billing_usage_snapshot_read_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_base_offering_version_id uuid,
    p_subscription_id uuid,
    p_subscription_offering_version_id uuid
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
), periods AS MATERIALIZED (
    SELECT period.*, total.committed_total, total.reserved_total,
           total.gauge_total, total.revision AS total_revision,
           total.updated_at,
           pg_catalog.count(*) OVER (
               PARTITION BY period.entitlement_key
           ) AS period_count
    FROM runtime_guard
    CROSS JOIN platform.billing_meter_periods AS period
    JOIN platform.billing_usage_totals AS total
      ON total.authority_id = period.authority_id
     AND total.tenant_id = period.tenant_id
     AND total.meter_period_id = period.id
    WHERE period.authority_id = p_authority_id
      AND period.tenant_id = p_tenant_id
      AND period.state = 'open'
      AND period.period_start <= pg_catalog.transaction_timestamp()
      AND period.period_end > pg_catalog.transaction_timestamp()
)
SELECT period.period_count,
       COALESCE(override_row.override_count, 0::bigint),
       period.id, period.entitlement_key, definition.value_type,
       definition.meter_kind, definition.reset_policy,
       definition.enforcement, definition.unlimited_allowed,
       period.period_start, period.period_end, period.revision,
       period.committed_total, period.reserved_total, period.gauge_total,
       period.total_revision, period.updated_at,
       CASE
         WHEN override_row.id IS NOT NULL THEN override_row.integer_value
         WHEN purchased.entitlement_definition_id IS NOT NULL
           THEN purchased.integer_value
         ELSE base_grant.integer_value
       END,
       CASE
         WHEN override_row.id IS NOT NULL THEN override_row.unlimited
         WHEN purchased.entitlement_definition_id IS NOT NULL
           THEN purchased.unlimited
         ELSE base_grant.unlimited
       END,
       (override_row.id IS NOT NULL
        OR purchased.entitlement_definition_id IS NOT NULL
        OR base_grant.entitlement_definition_id IS NOT NULL)
FROM periods AS period
JOIN platform.billing_entitlement_definitions AS definition
  ON definition.id = period.entitlement_definition_id
 AND definition.entitlement_key = period.entitlement_key
 AND definition.value_type = 'integer'
 AND definition.meter_kind = period.meter_kind
LEFT JOIN platform.billing_offering_entitlements AS base_grant
  ON base_grant.offering_version_id = p_base_offering_version_id
 AND base_grant.entitlement_definition_id = definition.id
LEFT JOIN platform.billing_offering_entitlements AS purchased
  ON purchased.offering_version_id = p_subscription_offering_version_id
 AND purchased.entitlement_definition_id = definition.id
LEFT JOIN LATERAL (
    SELECT candidate.*, pg_catalog.count(*) OVER () AS override_count
    FROM platform.billing_subscription_overrides AS candidate
    WHERE p_subscription_id IS NOT NULL
      AND candidate.authority_id = p_authority_id
      AND candidate.tenant_id = p_tenant_id
      AND candidate.subscription_id = p_subscription_id
      AND candidate.entitlement_definition_id = definition.id
      AND candidate.effective_start <= pg_catalog.transaction_timestamp()
      AND (candidate.effective_end IS NULL
           OR candidate.effective_end > pg_catalog.transaction_timestamp())
    ORDER BY candidate.effective_start DESC, candidate.id
    LIMIT 1
) AS override_row ON true
ORDER BY period.entitlement_key COLLATE "C", period.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.billing_usage_context_snapshot_read_v1(uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.billing_usage_snapshot_read_v1(uuid,uuid,uuid,uuid,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.billing_usage_context_snapshot_read_v1(uuid,uuid)
    TO harpy_runtime;
GRANT EXECUTE ON FUNCTION
    platform.billing_usage_snapshot_read_v1(uuid,uuid,uuid,uuid,uuid)
    TO harpy_runtime;
