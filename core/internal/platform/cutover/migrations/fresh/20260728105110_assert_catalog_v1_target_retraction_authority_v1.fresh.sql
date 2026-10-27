-- +goose Up
-- Migration stream version: 20260728105110.

SELECT 1 / CASE WHEN EXISTS (
    SELECT 1
    FROM platform.domain_cutovers AS domain
    LEFT JOIN platform.cutover_groups AS authority
      ON authority.scope_kind = domain.scope_kind
     AND authority.scope_id = domain.scope_id
     AND authority.group_id = 'catalog'
    LEFT JOIN platform.cutover_group_members AS member
      ON member.scope_kind = authority.scope_kind
     AND member.scope_id = authority.scope_id
     AND member.group_id = authority.group_id
     AND member.ordinal = 1
     AND member.domain = 'catalog'
    WHERE domain.scope_kind = 'tenant'
      AND domain.domain = 'catalog'
      AND (
          authority.scope_id IS NULL
          OR member.scope_id IS NULL
          OR authority.state IS DISTINCT FROM domain.state
          OR authority.writer IS DISTINCT FROM domain.writer
          OR authority.generation IS DISTINCT FROM domain.generation
          OR domain.generation < 1
          OR (
              SELECT pg_catalog.count(*)
              FROM platform.cutover_group_members AS sibling
              WHERE sibling.scope_kind = domain.scope_kind
                AND sibling.scope_id = domain.scope_id
                AND sibling.group_id = 'catalog'
          ) <> 1
          OR NOT (
              (domain.state = 'v2' AND domain.writer = 'v2')
              OR (
                  domain.state IN ('v1', 'shadow')
                  AND domain.writer = 'v1'
              )
              OR (
                  domain.state IN ('quiescing', 'fenced')
                  AND domain.writer = 'none'
              )
          )
      )
) THEN 0 ELSE 1 END;
