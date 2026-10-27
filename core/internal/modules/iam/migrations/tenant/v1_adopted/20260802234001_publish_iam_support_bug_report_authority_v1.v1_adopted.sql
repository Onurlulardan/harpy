-- +goose Up
MERGE INTO iam_resources AS target
USING (VALUES (
    '019fb201-0000-8000-8000-000000000001'::uuid,
    'bug_reports'::text, 'Support Bug Reports'::text
)) AS expected(id, resource_key, display_name)
ON target.resource_key = expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM 'support' OR
    target.resource_kind IS DISTINCT FROM 'support_bug_reports' OR
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, expected.resource_key, 'support', 'support_bug_reports',
    expected.display_name, 'Tenant-scoped Support bug-report authority.',
    true, 'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
    ('019fb201-0000-8000-8000-000000000101'::uuid, 'create'::text, 'Create'::text),
    ('019fb201-0000-8000-8000-000000000102', 'read_own', 'Read Own'),
    ('019fb201-0000-8000-8000-000000000103', 'read', 'Read'),
    ('019fb201-0000-8000-8000-000000000104', 'manage', 'Manage')
) AS expected(id, action_key, display_name)
ON target.action_key = expected.action_key
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, action_key, display_name, description, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, expected.action_key, expected.display_name, NULL,
    false, 'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_permissions AS target
USING (
    SELECT expected.id, resource.id AS resource_id, action.id AS action_id
    FROM (VALUES
        ('019fb201-0000-8000-8000-000000000201'::uuid, 'create'::text),
        ('019fb201-0000-8000-8000-000000000202', 'read_own'),
        ('019fb201-0000-8000-8000-000000000203', 'read'),
        ('019fb201-0000-8000-8000-000000000204', 'manage')
    ) AS expected(id, action_key)
    JOIN iam_resources AS resource ON resource.resource_key = 'bug_reports'
    JOIN iam_actions AS action ON action.action_key = expected.action_key
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

SELECT 1 / CASE WHEN (
    SELECT pg_catalog.count(*)
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
    JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE permission.match_kind = 'exact'
      AND permission.is_protected
      AND permission.state = 'active'
      AND resource.resource_key = 'bug_reports'
      AND resource.owner_domain = 'support'
      AND resource.resource_kind = 'support_bug_reports'
      AND resource.is_protected
      AND resource.state = 'active'
      AND action.action_key IN ('create', 'read_own', 'read', 'manage')
      AND NOT action.is_protected
      AND action.state = 'active'
) = 4 THEN 1 ELSE 0 END AS iam_support_bug_report_authority_assertion;
