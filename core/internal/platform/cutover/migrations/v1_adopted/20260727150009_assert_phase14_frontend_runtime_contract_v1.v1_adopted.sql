-- +goose Up
-- Migration stream version: 20260727150009.

SELECT 1 / CASE WHEN EXISTS (
    SELECT 1
    FROM platform.domain_cutovers AS domain
    WHERE domain.scope_kind = 'tenant'
      AND domain.domain IN ('analytics','iam','notifications')
      AND (domain.state <> 'v2' OR domain.writer <> 'v2')
) OR EXISTS (
    SELECT 1
    FROM platform.domain_cutovers AS domain
    WHERE domain.scope_kind = 'tenant'
      AND domain.domain = 'notifications'
      AND NOT EXISTS (
          SELECT 1
          FROM platform.cutover_groups AS authority
          JOIN platform.cutover_group_members AS member
            ON member.scope_kind = authority.scope_kind
           AND member.scope_id = authority.scope_id
           AND member.group_id = authority.group_id
          WHERE authority.scope_kind = domain.scope_kind
            AND authority.scope_id = domain.scope_id
            AND authority.group_id = 'notifications'
            AND authority.state = 'v2'
            AND authority.writer = 'v2'
            AND member.ordinal = 1
            AND member.domain = 'notifications'
      )
) THEN 0 ELSE 1 END;
