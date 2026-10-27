-- +goose Up
MERGE INTO iam_resources AS target
USING (VALUES
    ('fa100000-0000-4000-8000-000000000001'::uuid,
     'forms.definition'::text, 'Forms Definitions'::text),
    ('fa100000-0000-4000-8000-000000000002',
     'forms.publication', 'Forms Publications')
) AS expected(id, resource_key, display_name)
ON target.resource_key = expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM 'forms' OR
    target.resource_kind IS DISTINCT FROM 'system_resource' OR
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, expected.resource_key, 'forms', 'system_resource',
    expected.display_name, NULL, true, 'active', 1,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
    ('fa100000-0000-4000-8000-000000000101'::uuid, 'create'::text, 'Create'::text),
    ('fa100000-0000-4000-8000-000000000102', 'read', 'Read'),
    ('fa100000-0000-4000-8000-000000000103', 'update', 'Update'),
    ('fa100000-0000-4000-8000-000000000104', 'retire', 'Retire'),
    ('fa100000-0000-4000-8000-000000000105', 'validate', 'Validate'),
    ('fa100000-0000-4000-8000-000000000106', 'plan', 'Plan'),
    ('fa100000-0000-4000-8000-000000000107', 'activate', 'Activate'),
    ('fa100000-0000-4000-8000-000000000108', 'deactivate', 'Deactivate')
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
        ('fa100000-0000-4000-8000-000000000201'::uuid,
         'forms.definition'::text, 'read'::text),
        ('fa100000-0000-4000-8000-000000000202',
         'forms.definition', 'create'),
        ('fa100000-0000-4000-8000-000000000203',
         'forms.definition', 'update'),
        ('fa100000-0000-4000-8000-000000000204',
         'forms.definition', 'retire'),
        ('fa100000-0000-4000-8000-000000000205',
         'forms.publication', 'read'),
        ('fa100000-0000-4000-8000-000000000206',
         'forms.publication', 'create'),
        ('fa100000-0000-4000-8000-000000000207',
         'forms.publication', 'validate'),
        ('fa100000-0000-4000-8000-000000000208',
         'forms.publication', 'plan'),
        ('fa100000-0000-4000-8000-000000000209',
         'forms.publication', 'activate'),
        ('fa100000-0000-4000-8000-000000000210',
         'forms.publication', 'deactivate')
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
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, match_kind, resource_id, action_id, is_protected,
    state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, 'exact', expected.resource_id, expected.action_id, false,
    'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

UPDATE iam_permissions AS permission SET revision = permission.revision
FROM iam_resources AS resource, iam_actions AS action
WHERE permission.resource_id = resource.id
  AND permission.action_id = action.id
  AND permission.match_kind = 'exact'
  AND (
    resource.resource_key = 'forms.definition'
    AND action.action_key NOT IN ('read', 'create', 'update', 'retire')
    OR resource.resource_key = 'forms.publication'
    AND action.action_key NOT IN (
        'read', 'create', 'validate', 'plan', 'activate', 'deactivate'
    )
  );

