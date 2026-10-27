-- +goose Up
MERGE INTO iam_resources AS target
USING (VALUES (
    '019f9e00-0000-8000-8000-000000000001'::uuid,
    'iam.capability_grant'::text,
    'iam'::text,
    'capability_grant'::text,
    'IAM Capability Grants'::text
)) AS expected(id, resource_key, owner_domain, resource_kind, display_name)
ON target.resource_key = expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM expected.owner_domain OR
    target.resource_kind IS DISTINCT FROM expected.resource_kind OR
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, expected.resource_key, expected.owner_domain,
    expected.resource_kind, expected.display_name,
    'Controls tenant-local service-principal capability grants.',
    true, 'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_permissions AS target
USING (
    SELECT
        '019f9e00-0000-8000-8000-000000000201'::uuid AS id,
        resource.id AS resource_id,
        action.id AS action_id
    FROM iam_resources AS resource
    JOIN iam_actions AS action ON action.action_key = 'manage'
    WHERE resource.resource_key = 'iam.capability_grant'
      AND resource.owner_domain = 'iam'
      AND resource.resource_kind = 'capability_grant'
      AND resource.is_protected
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
      AND permission.is_protected
      AND permission.state = 'active'
      AND resource.resource_key = 'iam.capability_grant'
      AND resource.owner_domain = 'iam'
      AND resource.resource_kind = 'capability_grant'
      AND resource.is_protected
      AND resource.state = 'active'
      AND action.action_key = 'manage'
      AND action.state = 'active'
) = 1 THEN 1 ELSE 0 END
AS iam_workflow_publication_authority_assertion;
