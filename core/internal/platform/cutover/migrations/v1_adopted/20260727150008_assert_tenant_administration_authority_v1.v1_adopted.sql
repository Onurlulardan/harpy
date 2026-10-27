-- +goose Up
-- Migration stream version: 20260727150008.

SELECT 1 / CASE WHEN EXISTS (
    SELECT 1
    FROM platform.domain_cutovers AS iam
    LEFT JOIN platform.domain_cutovers AS catalog
      ON catalog.scope_kind = iam.scope_kind
     AND catalog.scope_id = iam.scope_id
     AND catalog.domain = 'catalog'
    LEFT JOIN platform.domain_cutovers AS billing
      ON billing.scope_kind = iam.scope_kind
     AND billing.scope_id = iam.scope_id
     AND billing.domain = 'billing'
    LEFT JOIN platform.domain_cutovers AS documents
      ON documents.scope_kind = iam.scope_kind
     AND documents.scope_id = iam.scope_id
     AND documents.domain = 'documents'
    LEFT JOIN platform.cutover_groups AS authority
      ON authority.scope_kind = iam.scope_kind
     AND authority.scope_id = iam.scope_id
     AND authority.group_id = 'iam_system_resources'
    LEFT JOIN platform.cutover_group_members AS iam_member
      ON iam_member.scope_kind = authority.scope_kind
     AND iam_member.scope_id = authority.scope_id
     AND iam_member.group_id = authority.group_id
     AND iam_member.ordinal = 1
     AND iam_member.domain = 'iam'
    LEFT JOIN platform.cutover_group_members AS layout_member
      ON layout_member.scope_kind = authority.scope_kind
     AND layout_member.scope_id = authority.scope_id
     AND layout_member.group_id = authority.group_id
     AND layout_member.ordinal = 2
     AND layout_member.domain = 'system_resource_layout'
    WHERE iam.scope_kind = 'tenant'
      AND iam.domain = 'iam'
      AND (
          iam.state <> 'v2'
          OR iam.writer <> 'v2'
          OR catalog.scope_id IS NULL
          OR catalog.state IS DISTINCT FROM iam.state
          OR catalog.writer IS DISTINCT FROM iam.writer
          OR billing.scope_id IS NULL
          OR billing.state IS DISTINCT FROM iam.state
          OR billing.writer IS DISTINCT FROM iam.writer
          OR documents.scope_id IS NULL
          OR documents.state IS DISTINCT FROM iam.state
          OR documents.writer IS DISTINCT FROM iam.writer
          OR authority.scope_id IS NULL
          OR authority.state IS DISTINCT FROM iam.state
          OR authority.writer IS DISTINCT FROM iam.writer
          OR iam_member.scope_id IS NULL
          OR layout_member.scope_id IS NULL
      )
) THEN 0 ELSE 1 END;
