-- +goose Up
-- Matching semantic keys preserve adopted RFC UUID identities.
-- Any owner, kind, protection, state, or permission-shape drift aborts.
MERGE INTO iam_resources AS target
USING (VALUES (
    '019c0000-0418-7000-8000-000000000001'::uuid,
    'workflow.instance'::text
)) AS expected(id, resource_key)
ON target.resource_key = expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM 'workflow' OR
    target.resource_kind IS DISTINCT FROM 'workflow_instance' OR
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active' OR
    pg_catalog.substr(target.id::text,20,1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
    id,resource_key,owner_domain,resource_kind,display_name,description,
    is_protected,state,revision,created_at,updated_at,retired_at
) VALUES (
    expected.id,expected.resource_key,'workflow','workflow_instance',
    'Workflow Instances',NULL,false,'active',1,
    pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp(),NULL
);

MERGE INTO iam_actions AS target
USING (VALUES (
    '019c0000-0418-7000-8000-000000000101'::uuid,
    'start'::text,
    'Start'::text
)) AS expected(id,action_key,display_name)
ON target.action_key=expected.action_key
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active' OR
    pg_catalog.substr(target.id::text,20,1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
    id,action_key,display_name,description,is_protected,
    state,revision,created_at,updated_at,retired_at
) VALUES (
    expected.id,expected.action_key,expected.display_name,NULL,false,
    'active',1,pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(),NULL
);

MERGE INTO iam_permissions AS target
USING (
    SELECT
        '019c0000-0418-7000-8000-000000000201'::uuid AS id,
        resource.id AS resource_id,
        action.id AS action_id
    FROM iam_resources AS resource
    JOIN iam_actions AS action ON action.action_key='start'
    WHERE resource.resource_key='workflow.instance'
) AS expected
ON target.match_kind='exact'
AND target.resource_id=expected.resource_id
AND target.action_id=expected.action_id
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active' OR
    pg_catalog.substr(target.id::text,20,1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
    id,match_kind,resource_id,action_id,is_protected,
    state,revision,created_at,updated_at,retired_at
) VALUES (
    expected.id,'exact',expected.resource_id,expected.action_id,false,
    'active',1,pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(),NULL
);

