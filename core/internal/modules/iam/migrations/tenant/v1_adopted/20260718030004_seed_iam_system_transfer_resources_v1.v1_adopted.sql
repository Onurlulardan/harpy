-- +goose Up
MERGE INTO iam_resources AS target
USING (VALUES
    ('d1030000-0000-4000-8000-000000000001'::uuid, 'iam.users'::text,
     'IAM Users'::text),
    ('d1030000-0000-4000-8000-000000000002', 'iam.roles', 'IAM Roles'),
    ('d1030000-0000-4000-8000-000000000003', 'iam.organizations',
     'IAM Organizations')
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
    ('d1030000-0000-4000-8000-000000000101'::uuid, 'import'::text,
     'Import'::text),
    ('d1030000-0000-4000-8000-000000000102', 'export', 'Export')
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

SELECT 1 / CASE WHEN
    (SELECT pg_catalog.count(*) FROM iam_resources
     WHERE resource_key IN ('iam.users','iam.roles','iam.organizations')
       AND owner_domain = 'iam' AND resource_kind = 'system_resource'
       AND is_protected AND state = 'active') = 3
    AND
    (SELECT pg_catalog.count(*) FROM iam_actions
     WHERE action_key IN ('import','export')
       AND NOT is_protected AND state = 'active') = 2
THEN 1 ELSE 0 END AS iam_system_transfer_authority_assertion;
