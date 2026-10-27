-- +goose Up
-- Semantic-key matches preserve adopted identities; incompatible drift aborts.
MERGE INTO iam_resources AS target
USING (VALUES
    ('d1030000-0000-4000-8000-000000000001'::uuid,
     'iam.users'::text, 'IAM Users'::text),
    ('d1030000-0000-4000-8000-000000000002',
     'iam.roles', 'IAM Roles'),
    ('d1030000-0000-4000-8000-000000000003',
     'iam.organizations', 'IAM Organizations'),
    ('019f9c00-0000-8000-8000-000000000001',
     'iam.permissions', 'IAM Permissions'),
    ('019f9c00-0000-8000-8000-000000000002',
     'iam.permission_requests', 'IAM Permission Requests')
) AS expected(id, resource_key, display_name)
ON target.resource_key = expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM 'iam' OR
    target.resource_kind IS DISTINCT FROM 'system_resource' OR
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, expected.resource_key, 'iam', 'system_resource',
    expected.display_name, NULL, true, 'active', 1,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
    ('019f9c00-0000-8000-8000-000000000101'::uuid, 'create'::text, 'Create'::text),
    ('019f9c00-0000-8000-8000-000000000102', 'read', 'Read'),
    ('019f9c00-0000-8000-8000-000000000103', 'update', 'Update'),
    ('019f9c00-0000-8000-8000-000000000104', 'delete', 'Delete'),
    ('019f9c00-0000-8000-8000-000000000105', 'manage', 'Manage')
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
        ('019f9c00-0000-8000-8000-000000000201'::uuid,
         'iam.users'::text, 'create'::text),
        ('019f9c00-0000-8000-8000-000000000202', 'iam.users', 'read'),
        ('019f9c00-0000-8000-8000-000000000203', 'iam.users', 'update'),
        ('019f9c00-0000-8000-8000-000000000204', 'iam.users', 'delete'),
        ('019f9c00-0000-8000-8000-000000000205', 'iam.roles', 'create'),
        ('019f9c00-0000-8000-8000-000000000206', 'iam.roles', 'read'),
        ('019f9c00-0000-8000-8000-000000000207', 'iam.roles', 'update'),
        ('019f9c00-0000-8000-8000-000000000208', 'iam.roles', 'delete'),
        ('019f9c00-0000-8000-8000-000000000209', 'iam.organizations', 'create'),
        ('019f9c00-0000-8000-8000-000000000210', 'iam.organizations', 'read'),
        ('019f9c00-0000-8000-8000-000000000211', 'iam.organizations', 'update'),
        ('019f9c00-0000-8000-8000-000000000212', 'iam.organizations', 'delete'),
        ('019f9c00-0000-8000-8000-000000000213', 'iam.permissions', 'create'),
        ('019f9c00-0000-8000-8000-000000000214', 'iam.permissions', 'read'),
        ('019f9c00-0000-8000-8000-000000000215', 'iam.permissions', 'update'),
        ('019f9c00-0000-8000-8000-000000000216', 'iam.permissions', 'delete'),
        ('019f9c00-0000-8000-8000-000000000217',
         'iam.permission_requests', 'read'),
        ('019f9c00-0000-8000-8000-000000000218',
         'iam.permission_requests', 'manage')
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
    SELECT pg_catalog.count(*)
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
    JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE permission.match_kind = 'exact'
      AND permission.is_protected AND permission.state = 'active'
      AND resource.resource_key IN (
          'iam.users', 'iam.roles', 'iam.organizations', 'iam.permissions',
          'iam.permission_requests'
      )
      AND action.action_key IN ('create', 'read', 'update', 'delete', 'manage')
) = 18 THEN 1 ELSE 0 END AS iam_operator_admin_authority_assertion;
