-- +goose Up
CREATE VIEW platform.billing_discount_query_v1
WITH (security_barrier = true)
AS
WITH current_authority AS (
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
SELECT version_row.id,
       root.id AS discount_id,
       root.normalized_code AS code,
       root.normalized_code AS code_folded,
       root.revision,
       root.lifecycle,
       version_row.version,
       version_row.discount_kind,
       version_row.fixed_minor_units,
       version_row.currency::text AS currency,
       version_row.currency_registry_version,
       version_row.basis_points,
       version_row.duration,
       version_row.repeating_cycles,
       version_row.priority,
       version_row.global_use_limit,
       version_row.per_tenant_use_limit,
       version_row.all_offerings,
       version_row.expected_offering_count,
       version_row.valid_start,
       version_row.valid_end,
       version_row.published_at
FROM current_authority AS authority
JOIN platform.billing_discounts AS root
  ON root.authority_id = authority.authority_id
JOIN platform.billing_discount_versions AS version_row
  ON version_row.authority_id = authority.authority_id
 AND version_row.discount_id = root.id
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';

GRANT SELECT ON platform.billing_discount_query_v1 TO harpy_runtime;
