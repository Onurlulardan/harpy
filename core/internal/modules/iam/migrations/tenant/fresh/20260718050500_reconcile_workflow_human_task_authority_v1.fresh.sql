-- +goose Up
-- A drift match intentionally violates the revision-step guard and aborts.
MERGE INTO iam_resources AS target
USING (VALUES (
    '019c0000-0500-7000-8000-000000000001'::uuid,
    'workflow.human_tasks'::text
)) AS expected(id, resource_key)
ON target.resource_key = expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM 'workflow' OR
    target.resource_kind IS DISTINCT FROM 'human_task' OR
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active' OR
    pg_catalog.substr(target.id::text, 20, 1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, resource_key, owner_domain, resource_kind, display_name, description,
    is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, expected.resource_key, 'workflow', 'human_task',
    'Workflow Human Tasks', NULL, false, 'active', 1,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
    ('019c0000-0500-7000-8000-000000000101'::uuid, 'claim'::text, 'Claim'::text),
    ('019c0000-0500-7000-8000-000000000102', 'decide', 'Decide'),
    ('019c0000-0500-7000-8000-000000000103', 'reassign', 'Reassign'),
    ('019c0000-0500-7000-8000-000000000104', 'expire', 'Expire'),
    ('019c0000-0500-7000-8000-000000000105', 'cancel', 'Cancel')
) AS expected(id, action_key, display_name)
ON target.action_key = expected.action_key
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active' OR
    pg_catalog.substr(target.id::text, 20, 1) NOT IN ('8','9','a','b')
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
        ('019c0000-0500-7000-8000-000000000201'::uuid, 'claim'::text),
        ('019c0000-0500-7000-8000-000000000202', 'decide'),
        ('019c0000-0500-7000-8000-000000000203', 'reassign'),
        ('019c0000-0500-7000-8000-000000000204', 'expire'),
        ('019c0000-0500-7000-8000-000000000205', 'cancel')
    ) AS expected(id, action_key)
    JOIN iam_resources AS resource
      ON resource.resource_key = 'workflow.human_tasks'
    JOIN iam_actions AS action ON action.action_key = expected.action_key
) AS expected
ON target.match_kind = 'exact'
AND target.resource_id = expected.resource_id
AND target.action_id = expected.action_id
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active' OR
    pg_catalog.substr(target.id::text, 20, 1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, 'exact', expected.resource_id, expected.action_id, false,
    'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);
