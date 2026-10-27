-- +goose Up
CREATE VIEW platform.billing_offering_query_v1
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
      AND group_row.writer = 'v2'
      AND group_row.state = 'v2'
      AND domain_row.writer = 'v2'
      AND domain_row.state = 'v2'
)
SELECT
    version_row.id,
    root.id AS offering_id,
    root.offering_key,
    root.offering_key AS offering_key_folded,
    pg_catalog.normalize(
        version_row.presentation_metadata ->> 'name', 'NFC'
    ) AS name,
    root.revision,
    root.lifecycle,
    version_row.version,
    version_row.is_base,
    version_row.is_public,
    version_row.trial_days,
    version_row.grace_days,
    version_row.expected_price_count,
    version_row.expected_grant_count,
    version_row.effective_start,
    version_row.effective_end,
    version_row.publication_state,
    version_row.published_at
FROM current_authority AS authority
JOIN platform.billing_offerings AS root
  ON root.authority_id = authority.authority_id
JOIN platform.billing_offering_versions AS version_row
  ON version_row.authority_id = authority.authority_id
 AND version_row.offering_id = root.id
WHERE SESSION_USER IS NOT DISTINCT FROM 'harpy_v2_runtime'
  AND CURRENT_USER IS NOT DISTINCT FROM 'harpy_runtime';

GRANT SELECT ON platform.billing_offering_query_v1 TO harpy_runtime;
