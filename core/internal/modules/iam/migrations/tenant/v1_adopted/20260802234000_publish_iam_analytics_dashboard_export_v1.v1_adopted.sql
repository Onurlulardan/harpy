-- +goose Up
MERGE INTO iam_actions AS target
USING (VALUES (
    '019fb200-0000-8000-8000-000000000101'::uuid,
    'export'::text, 'Export'::text
)) AS expected(id, action_key, display_name)
ON target.action_key = expected.action_key
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, action_key, display_name, description, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, expected.action_key, expected.display_name,
    'Export an authorized dashboard snapshot.', false, 'active', 1,
    pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_permissions AS target
USING (
    SELECT '019fb200-0000-8000-8000-000000000201'::uuid AS id,
           resource.id AS resource_id, action.id AS action_id
    FROM iam_resources AS resource
    JOIN iam_actions AS action ON action.action_key = 'export'
    WHERE resource.resource_key = 'analytics.dashboard'
      AND resource.owner_domain = 'analytics'
      AND resource.state = 'active'
      AND action.state = 'active'
) AS expected
ON target.match_kind = 'exact'
AND target.resource_id = expected.resource_id
AND target.action_id = expected.action_id
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, 'exact', expected.resource_id, expected.action_id,
    true, 'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

SELECT 1 / CASE WHEN EXISTS (
    SELECT 1
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
    JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE permission.match_kind = 'exact'
      AND permission.is_protected
      AND permission.state = 'active'
      AND resource.resource_key = 'analytics.dashboard'
      AND resource.owner_domain = 'analytics'
      AND resource.state = 'active'
      AND action.action_key = 'export'
      AND NOT action.is_protected
      AND action.state = 'active'
) THEN 1 ELSE 0 END AS iam_analytics_dashboard_export_authority_assertion;
