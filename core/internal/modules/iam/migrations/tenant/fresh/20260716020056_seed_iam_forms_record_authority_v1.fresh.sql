-- +goose Up
ALTER TABLE iam_permission_implications
    DROP CONSTRAINT iam_permission_implications_kind_check;
ALTER TABLE iam_permission_implications
    ADD CONSTRAINT iam_permission_implications_kind_check CHECK (
        implication_kind IN ('v1_form_parent_v1', 'forms_record_operation_v1')
    );

MERGE INTO iam_resources AS target
USING (VALUES ('f6000000-0000-4000-8000-000000000001'::uuid)) AS expected(id)
ON target.resource_key = 'forms.records'
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM 'forms' OR
    target.resource_kind IS DISTINCT FROM 'system_resource' OR
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET display_name = target.display_name
WHEN NOT MATCHED THEN INSERT (
    id, resource_key, owner_domain, resource_kind, display_name,
    description, is_protected, state, revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, 'forms.records', 'forms', 'system_resource', 'Forms Records',
    NULL, true, 'active', 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
    ('f6000000-0000-4000-8000-000000000101'::uuid, 'create'::text, 'Create'::text),
    ('f6000000-0000-4000-8000-000000000102', 'read', 'Read'),
    ('f6000000-0000-4000-8000-000000000103', 'update', 'Update'),
    ('f6000000-0000-4000-8000-000000000104', 'delete', 'Delete'),
    ('f6000000-0000-4000-8000-000000000105', 'import', 'Import'),
    ('f6000000-0000-4000-8000-000000000106', 'export', 'Export')
) AS expected(id, action_key, display_name)
ON target.action_key = expected.action_key
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET display_name = target.display_name
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
        ('f6000000-0000-4000-8000-000000000201'::uuid, 'create'::text),
        ('f6000000-0000-4000-8000-000000000202', 'read'),
        ('f6000000-0000-4000-8000-000000000203', 'update'),
        ('f6000000-0000-4000-8000-000000000204', 'delete'),
        ('f6000000-0000-4000-8000-000000000205', 'import'),
        ('f6000000-0000-4000-8000-000000000206', 'export')
    ) AS expected(id, action_key)
    JOIN iam_actions AS action ON action.action_key = expected.action_key
    JOIN iam_resources AS resource ON resource.resource_key = 'forms.records'
) AS expected
ON target.match_kind = 'exact'
AND target.resource_id = expected.resource_id
AND target.action_id = expected.action_id
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active'
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
WHERE permission.resource_id = resource.id AND permission.action_id = action.id
  AND permission.match_kind = 'exact' AND resource.resource_key = 'forms.records'
  AND action.action_key NOT IN ('create', 'read', 'update', 'delete', 'import', 'export');

MERGE INTO iam_permission_implications AS target
USING (
    SELECT expected.id, source.id AS source_id, implied.id AS implied_id
    FROM (VALUES
        ('f6000000-0000-4000-8000-000000000301'::uuid,
         'import'::text, 'create'::text),
        ('f6000000-0000-4000-8000-000000000302', 'export', 'read')
    ) AS expected(id, source_action, implied_action)
    JOIN iam_actions AS source_action
      ON source_action.action_key = expected.source_action
    JOIN iam_actions AS implied_action
      ON implied_action.action_key = expected.implied_action
    JOIN iam_resources AS resource ON resource.resource_key = 'forms.records'
    JOIN iam_permissions AS source
      ON source.match_kind = 'exact' AND source.resource_id = resource.id
     AND source.action_id = source_action.id
    JOIN iam_permissions AS implied
      ON implied.match_kind = 'exact' AND implied.resource_id = resource.id
     AND implied.action_id = implied_action.id
) AS expected
ON target.source_permission_id = expected.source_id
AND target.implied_permission_id = expected.implied_id
WHEN MATCHED AND (
    target.implication_kind IS DISTINCT FROM 'forms_record_operation_v1' OR
    target.source_match_kind IS DISTINCT FROM 'exact' OR
    target.implied_match_kind IS DISTINCT FROM 'exact' OR
    target.retirement_contract_key IS DISTINCT FROM 'forms.records.operation.v1' OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision = target.revision
WHEN NOT MATCHED THEN INSERT (
    id, implication_kind, source_permission_id, source_match_kind,
    implied_permission_id, implied_match_kind, retirement_contract_key,
    state, hit_count, last_hit_at, retirement_receipt_id,
    revision, created_at, updated_at, retired_at
) VALUES (
    expected.id, 'forms_record_operation_v1', expected.source_id, 'exact',
    expected.implied_id, 'exact', 'forms.records.operation.v1',
    'active', 0, NULL, NULL, 1, pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(), NULL
);

UPDATE iam_permission_implications AS edge SET revision = edge.revision
WHERE edge.implication_kind = 'forms_record_operation_v1'
  AND NOT EXISTS (
    SELECT 1 FROM iam_permissions AS source
    JOIN iam_actions AS source_action ON source_action.id = source.action_id
    JOIN iam_permissions AS implied ON implied.id = edge.implied_permission_id
    JOIN iam_actions AS implied_action ON implied_action.id = implied.action_id
    JOIN iam_resources AS resource ON resource.id = source.resource_id
    WHERE source.id = edge.source_permission_id
      AND implied.resource_id = resource.id
      AND resource.resource_key = 'forms.records'
      AND ((source_action.action_key = 'import' AND implied_action.action_key = 'create')
        OR (source_action.action_key = 'export' AND implied_action.action_key = 'read'))
  );
