-- +goose Up
INSERT INTO platform.billing_global_authorities (
    authority_id, scope_kind, scope_id, authority_group_id,
    authority_group_generation, authority_domain,
    authority_domain_generation, created_at
)
SELECT
    pg_catalog.gen_random_uuid(), 'global', environment.environment_id,
    'billing', group_row.generation, 'billing', domain_row.generation,
    pg_catalog.statement_timestamp()
FROM platform.environment_identity AS environment
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = 'global'
 AND group_row.scope_id = environment.environment_id
 AND group_row.group_id = 'billing'
JOIN platform.cutover_group_members AS member_row
  ON member_row.scope_kind = group_row.scope_kind
 AND member_row.scope_id = group_row.scope_id
 AND member_row.group_id = group_row.group_id
 AND member_row.domain = 'billing'
JOIN platform.domain_cutovers AS domain_row
  ON domain_row.scope_kind = member_row.scope_kind
 AND domain_row.scope_id = member_row.scope_id
 AND domain_row.domain = member_row.domain
WHERE group_row.state = domain_row.state
  AND group_row.writer = domain_row.writer;

SELECT CASE
    WHEN pg_catalog.count(*) = 1 THEN pg_catalog.count(*)
    ELSE 1 / (pg_catalog.count(*) - pg_catalog.count(*))
END AS billing_global_authority_cardinality
FROM platform.billing_global_authorities;
