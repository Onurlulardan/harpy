-- +goose Up
-- Classified IAM fields are part of the published system-resource catalog.
-- Keep their actions addressable so protected tenant-wide authority can
-- reproduce the exact catalog evidence pinned by published layouts.
MERGE INTO iam_actions AS target
USING (VALUES
    ('019fae00-0000-8000-8000-000000000101'::uuid,
     'read.confidential'::text, 'Read Confidential'::text),
    ('019fae00-0000-8000-8000-000000000102',
     'read.restricted', 'Read Restricted')
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
    expected.id, expected.action_key, expected.display_name,
    'Controls classified IAM system-resource field visibility.',
    false, 'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_permissions AS target
USING (
    SELECT expected.id, resource.id AS resource_id, action.id AS action_id
    FROM (VALUES
        ('019fae00-0000-8000-8000-000000000201'::uuid,
         'iam.users'::text, 'read.confidential'::text),
        ('019fae00-0000-8000-8000-000000000202',
         'iam.users', 'read.restricted'),
        ('019fae00-0000-8000-8000-000000000203',
         'iam.roles', 'read.confidential'),
        ('019fae00-0000-8000-8000-000000000204',
         'iam.roles', 'read.restricted'),
        ('019fae00-0000-8000-8000-000000000205',
         'iam.organizations', 'read.confidential'),
        ('019fae00-0000-8000-8000-000000000206',
         'iam.organizations', 'read.restricted')
    ) AS expected(id, resource_key, action_key)
    JOIN iam_resources AS resource
      ON resource.resource_key = expected.resource_key
     AND resource.owner_domain = 'iam'
     AND resource.resource_kind = 'system_resource'
     AND resource.is_protected
     AND resource.state = 'active'
    JOIN iam_actions AS action
      ON action.action_key = expected.action_key
     AND NOT action.is_protected
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
    expected.id, 'exact', expected.resource_id, expected.action_id, true,
    'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

SELECT 1 / CASE WHEN
    (SELECT pg_catalog.count(*)
     FROM iam_actions
     WHERE action_key IN ('read.confidential', 'read.restricted')
       AND NOT is_protected AND state = 'active') = 2
    AND
    (SELECT pg_catalog.count(*)
     FROM iam_permissions AS permission
     JOIN iam_resources AS resource ON resource.id = permission.resource_id
     JOIN iam_actions AS action ON action.id = permission.action_id
     WHERE permission.match_kind = 'exact'
       AND permission.is_protected
       AND permission.state = 'active'
       AND resource.resource_key IN (
           'iam.users', 'iam.roles', 'iam.organizations'
       )
       AND resource.owner_domain = 'iam'
       AND resource.resource_kind = 'system_resource'
       AND resource.is_protected
       AND resource.state = 'active'
       AND action.action_key IN ('read.confidential', 'read.restricted')
       AND NOT action.is_protected
       AND action.state = 'active') = 6
THEN 1 ELSE 0 END AS iam_field_classification_authority_assertion;
