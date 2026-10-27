-- +goose Up
MERGE INTO iam_resources AS target
USING (VALUES
    ('019fb300-0000-8000-8000-000000000001'::uuid,
     'work.boards'::text, 'Work Boards'::text),
    ('019fb300-0000-8000-8000-000000000002',
     'work.columns', 'Work Columns'),
    ('019fb300-0000-8000-8000-000000000003',
     'work.memberships', 'Work Memberships'),
    ('019fb300-0000-8000-8000-000000000004',
     'work.tasks', 'Work Tasks'),
    ('019fb300-0000-8000-8000-000000000005',
     'work.comments', 'Work Comments'),
    ('019fb300-0000-8000-8000-000000000006',
     'work.labels', 'Work Labels'),
    ('019fb300-0000-8000-8000-000000000007',
     'work.checklists', 'Work Checklists'),
    ('019fb300-0000-8000-8000-000000000008',
     'work.calendar', 'Work Calendar'),
    ('019fb300-0000-8000-8000-000000000009',
     'work.recurrence', 'Work Recurrence'),
    ('019fb300-0000-8000-8000-00000000000a',
     'work.inbox', 'Work Inbox')
) AS expected(id, resource_key, display_name)
ON target.resource_key = expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM 'work' OR
    target.resource_kind IS DISTINCT FROM 'system_resource' OR
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, expected.resource_key, 'work', 'system_resource',
    expected.display_name, NULL, true, 'active', 1,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
    ('019fb300-0000-8000-8000-000000000101'::uuid, 'read'::text, 'Read'::text),
    ('019fb300-0000-8000-8000-000000000102', 'create', 'Create'),
    ('019fb300-0000-8000-8000-000000000103', 'update', 'Update'),
    ('019fb300-0000-8000-8000-000000000104', 'archive', 'Archive'),
    ('019fb300-0000-8000-8000-000000000105', 'restore', 'Restore'),
    ('019fb300-0000-8000-8000-000000000106', 'manage', 'Manage'),
    ('019fb300-0000-8000-8000-000000000107', 'move', 'Move'),
    ('019fb300-0000-8000-8000-000000000108', 'assign', 'Assign'),
    ('019fb300-0000-8000-8000-000000000109', 'transition', 'Transition'),
    ('019fb300-0000-8000-8000-00000000010a', 'redact', 'Redact')
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
    expected.id, expected.action_key, expected.display_name, NULL, false,
    'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_permissions AS target
USING (
    SELECT expected.id, resource.id AS resource_id, action.id AS action_id
    FROM (VALUES
        ('019fb300-0000-8000-8000-000000000201'::uuid,
         'work.boards'::text, 'read'::text),
        ('019fb300-0000-8000-8000-000000000202', 'work.boards', 'create'),
        ('019fb300-0000-8000-8000-000000000203', 'work.boards', 'update'),
        ('019fb300-0000-8000-8000-000000000204', 'work.boards', 'archive'),
        ('019fb300-0000-8000-8000-000000000205', 'work.boards', 'restore'),
        ('019fb300-0000-8000-8000-000000000206', 'work.columns', 'create'),
        ('019fb300-0000-8000-8000-000000000207', 'work.columns', 'update'),
        ('019fb300-0000-8000-8000-000000000208', 'work.memberships', 'manage'),
        ('019fb300-0000-8000-8000-000000000209', 'work.tasks', 'read'),
        ('019fb300-0000-8000-8000-00000000020a', 'work.tasks', 'create'),
        ('019fb300-0000-8000-8000-00000000020b', 'work.tasks', 'update'),
        ('019fb300-0000-8000-8000-00000000020c', 'work.tasks', 'move'),
        ('019fb300-0000-8000-8000-00000000020d', 'work.tasks', 'assign'),
        ('019fb300-0000-8000-8000-00000000020e', 'work.tasks', 'transition'),
        ('019fb300-0000-8000-8000-00000000020f', 'work.tasks', 'archive'),
        ('019fb300-0000-8000-8000-000000000210', 'work.tasks', 'restore'),
        ('019fb300-0000-8000-8000-000000000211', 'work.comments', 'read'),
        ('019fb300-0000-8000-8000-000000000212', 'work.comments', 'create'),
        ('019fb300-0000-8000-8000-000000000213', 'work.comments', 'redact'),
        ('019fb300-0000-8000-8000-000000000214', 'work.labels', 'manage'),
        ('019fb300-0000-8000-8000-000000000215', 'work.checklists', 'manage'),
        ('019fb300-0000-8000-8000-000000000216', 'work.calendar', 'read'),
        ('019fb300-0000-8000-8000-000000000217', 'work.calendar', 'manage'),
        ('019fb300-0000-8000-8000-000000000218', 'work.recurrence', 'manage'),
        ('019fb300-0000-8000-8000-000000000219', 'work.inbox', 'read')
    ) AS expected(id, resource_key, action_key)
    JOIN iam_resources AS resource
      ON resource.resource_key = expected.resource_key
    JOIN iam_actions AS action
      ON action.action_key = expected.action_key
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
    expected.id, 'exact', expected.resource_id, expected.action_id, true,
    'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

SELECT 1 / CASE WHEN (
    SELECT pg_catalog.count(*) FROM iam_resources
    WHERE owner_domain = 'work' AND resource_kind = 'system_resource'
      AND is_protected AND state = 'active'
      AND resource_key IN (
          'work.boards', 'work.columns', 'work.memberships', 'work.tasks',
          'work.comments', 'work.labels', 'work.checklists', 'work.calendar',
          'work.recurrence', 'work.inbox')
) = 10 AND (
    SELECT pg_catalog.count(*)
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
    JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE permission.match_kind = 'exact' AND permission.is_protected
      AND permission.state = 'active' AND resource.owner_domain = 'work'
      AND resource.resource_kind = 'system_resource' AND resource.is_protected
      AND resource.state = 'active' AND action.state = 'active'
      AND NOT action.is_protected
) = 25 THEN 1 ELSE 0 END AS iam_work_access_authority_assertion;
