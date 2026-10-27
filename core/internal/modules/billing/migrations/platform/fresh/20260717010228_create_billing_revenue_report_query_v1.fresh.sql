-- +goose Up
CREATE VIEW platform.billing_revenue_report_query_v1
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
)
SELECT
    projection.id,
    projection.state,
    projection.cadence,
    projection.currency,
    projection.currency_registry_version,
    projection.subscription_revision,
    projection.offering_version_id,
    projection.price_version_id,
    projection.monthly_minor_units,
    projection.annual_minor_units
FROM current_global AS authority
JOIN platform.billing_revenue_projection AS projection
  ON projection.authority_id = authority.authority_id
WHERE projection.state IN ('active', 'past_due')
  AND SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';

GRANT SELECT ON platform.billing_revenue_report_query_v1
    TO harpy_runtime;

