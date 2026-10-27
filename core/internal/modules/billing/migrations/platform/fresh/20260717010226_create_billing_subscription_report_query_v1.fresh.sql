-- +goose Up
CREATE VIEW platform.billing_subscription_report_query_v1
WITH (security_barrier = true)
AS
WITH current_global AS (
    SELECT authority.authority_id
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
    WHERE tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL')
      AND group_row.writer = 'v2' AND group_row.state = 'v2'
      AND billing_row.writer = 'v2' AND billing_row.state = 'v2'
      AND forms_row.writer = 'v2' AND forms_row.state = 'v2'
)
SELECT
    subscription.id,
    subscription.tenant_id,
    subscription.offering_version_id,
    subscription.price_version_id,
    provider.provider_key AS provider,
    subscription.state,
    subscription.period_start,
    subscription.period_end,
    subscription.trial_ends_at,
    subscription.grace_ends_at,
    subscription.cancel_at_period_end,
    subscription.revision AS version
FROM current_global
JOIN current_tenants AS authority ON TRUE
JOIN platform.billing_subscriptions AS subscription
  ON subscription.authority_id = authority.authority_id
 AND subscription.tenant_id = authority.tenant_id
JOIN platform.billing_provider_accounts AS provider
  ON provider.id = subscription.provider_account_id
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';

GRANT SELECT ON platform.billing_subscription_report_query_v1
    TO harpy_runtime;

