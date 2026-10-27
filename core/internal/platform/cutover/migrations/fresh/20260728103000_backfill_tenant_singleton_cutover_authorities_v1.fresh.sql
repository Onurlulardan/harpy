-- +goose Up
-- Migration stream version: 20260728103000.

WITH singleton(domain) AS (
    VALUES
        ('analytics'), ('catalog'), ('documents'),
        ('notifications'), ('work'), ('workflow')
)
INSERT INTO platform.cutover_groups (
    scope_kind, scope_id, group_id, state, writer, generation,
    requested_by, approved_by, reason, change_reference,
    acceptance_digest, changed_at
)
SELECT
    domain.scope_kind, domain.scope_id, singleton.domain,
    domain.state, domain.writer, domain.generation,
    domain.requested_by, domain.approved_by, domain.reason,
    domain.change_reference, domain.acceptance_digest, domain.changed_at
FROM platform.domain_cutovers AS domain
JOIN singleton ON singleton.domain = domain.domain
WHERE domain.scope_kind = 'tenant'
ON CONFLICT (scope_kind, scope_id, group_id) DO NOTHING;

INSERT INTO platform.cutover_group_members (
    scope_kind, scope_id, group_id, ordinal, domain
)
SELECT authority.scope_kind, authority.scope_id,
       authority.group_id, 1, authority.group_id
FROM platform.cutover_groups AS authority
WHERE authority.scope_kind = 'tenant'
  AND authority.group_id IN (
      'analytics', 'catalog', 'documents',
      'notifications', 'work', 'workflow'
  )
ON CONFLICT (scope_kind, scope_id, group_id, ordinal) DO NOTHING;

WITH tenants AS (
    SELECT DISTINCT scope_id
    FROM platform.domain_cutovers
    WHERE scope_kind = 'tenant'
), expected_group(group_id) AS (
    VALUES
        ('analytics'), ('billing'), ('catalog'), ('documents'),
        ('iam_system_resources'), ('notifications'), ('work'), ('workflow')
), expected_member(group_id, ordinal, domain) AS (
    VALUES
        ('analytics',1,'analytics'),
        ('billing',1,'billing'),
        ('catalog',1,'catalog'), ('documents',1,'documents'),
        ('iam_system_resources',1,'iam'),
        ('iam_system_resources',2,'system_resource_layout'),
        ('notifications',1,'notifications'), ('work',1,'work'),
        ('workflow',1,'workflow')
)
SELECT 1 / CASE WHEN EXISTS (
    SELECT 1 FROM tenants AS tenant
    WHERE (SELECT pg_catalog.count(*) FROM platform.domain_cutovers
           WHERE scope_kind = 'tenant' AND scope_id = tenant.scope_id) <> 9
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_groups
           WHERE scope_kind = 'tenant' AND scope_id = tenant.scope_id) <> 8
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
           WHERE scope_kind = 'tenant' AND scope_id = tenant.scope_id) <> 9
       OR (SELECT pg_catalog.count(*)
           FROM expected_group AS expected
           JOIN platform.cutover_groups AS authority
             ON authority.scope_kind = 'tenant'
            AND authority.scope_id = tenant.scope_id
            AND authority.group_id = expected.group_id) <> 8
       OR (SELECT pg_catalog.count(*)
           FROM expected_member AS expected
           JOIN platform.cutover_group_members AS member
             ON member.scope_kind = 'tenant'
            AND member.scope_id = tenant.scope_id
            AND member.group_id = expected.group_id
            AND member.ordinal = expected.ordinal
            AND member.domain = expected.domain) <> 9
) OR EXISTS (
    SELECT 1
    FROM platform.domain_cutovers AS domain
    JOIN platform.cutover_groups AS authority
      ON authority.scope_kind = domain.scope_kind
     AND authority.scope_id = domain.scope_id
     AND authority.group_id = domain.domain
    WHERE domain.scope_kind = 'tenant'
      AND domain.domain IN (
          'analytics', 'catalog', 'documents',
          'notifications', 'work', 'workflow'
      )
      AND (
          authority.state IS DISTINCT FROM domain.state
          OR authority.writer IS DISTINCT FROM domain.writer
          OR authority.generation IS DISTINCT FROM domain.generation
          OR authority.requested_by IS DISTINCT FROM domain.requested_by
          OR authority.approved_by IS DISTINCT FROM domain.approved_by
          OR authority.reason IS DISTINCT FROM domain.reason
          OR authority.change_reference IS DISTINCT FROM domain.change_reference
          OR authority.acceptance_digest IS DISTINCT FROM domain.acceptance_digest
      )
) THEN 0 ELSE 1 END;
