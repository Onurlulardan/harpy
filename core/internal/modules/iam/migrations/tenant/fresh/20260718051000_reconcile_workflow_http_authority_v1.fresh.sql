-- +goose Up
-- Semantic-key matches preserve adopted RFC UUID identities; drift aborts.
MERGE INTO iam_resources AS target
USING (VALUES
 ('019c0000-5100-7000-8000-000000000001'::uuid,'workflow.definition'::text,'workflow_definition'::text,'Workflow Definitions'::text),
 ('019c0000-5100-7000-8000-000000000002','workflow.publication','workflow_publication','Workflow Publications'),
 ('019c0000-0418-7000-8000-000000000001','workflow.instance','workflow_instance','Workflow Instances'),
 ('019c0000-0500-7000-8000-000000000001','workflow.human_tasks','human_task','Workflow Human Tasks')
) AS expected(id,resource_key,resource_kind,display_name)
ON target.resource_key=expected.resource_key
WHEN MATCHED AND (
 target.owner_domain IS DISTINCT FROM 'workflow' OR
 target.resource_kind IS DISTINCT FROM expected.resource_kind OR
 target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active' OR
 pg_catalog.substr(target.id::text,20,1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
 id,resource_key,owner_domain,resource_kind,display_name,description,
 is_protected,state,revision,created_at,updated_at,retired_at
) VALUES (
 expected.id,expected.resource_key,'workflow',expected.resource_kind,
 expected.display_name,NULL,false,'active',1,pg_catalog.transaction_timestamp(),
 pg_catalog.transaction_timestamp(),NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
 ('019c0000-5100-7000-8000-000000000101'::uuid,'create'::text,'Create'::text),
 ('019c0000-5100-7000-8000-000000000102','read','Read'),
 ('019c0000-5100-7000-8000-000000000103','update','Update'),
 ('019c0000-5100-7000-8000-000000000104','retire','Retire'),
 ('019c0000-5100-7000-8000-000000000105','publish','Publish'),
 ('019c0000-0418-7000-8000-000000000101','start','Start'),
 ('019c0000-5100-7000-8000-000000000106','signal','Signal'),
 ('019c0000-0500-7000-8000-000000000105','cancel','Cancel'),
 ('019c0000-5100-7000-8000-000000000107','replan','Replan'),
 ('019c0000-5100-7000-8000-000000000108','repair','Repair'),
 ('019c0000-0500-7000-8000-000000000101','claim','Claim'),
 ('019c0000-0500-7000-8000-000000000102','decide','Decide'),
 ('019c0000-0500-7000-8000-000000000103','reassign','Reassign'),
 ('019c0000-0500-7000-8000-000000000104','expire','Expire')
) AS expected(id,action_key,display_name)
ON target.action_key=expected.action_key
WHEN MATCHED AND (
 target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active' OR
 pg_catalog.substr(target.id::text,20,1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
 id,action_key,display_name,description,is_protected,state,revision,
 created_at,updated_at,retired_at
) VALUES (
 expected.id,expected.action_key,expected.display_name,NULL,false,'active',1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp(),NULL
);

MERGE INTO iam_permissions AS target
USING (
 SELECT expected.id,resource.id AS resource_id,action.id AS action_id
 FROM (VALUES
  ('019c0000-5100-7000-8000-000000000201'::uuid,'workflow.definition'::text,'create'::text),
  ('019c0000-5100-7000-8000-000000000202','workflow.definition','read'),
  ('019c0000-5100-7000-8000-000000000203','workflow.definition','update'),
  ('019c0000-5100-7000-8000-000000000204','workflow.definition','retire'),
  ('019c0000-5100-7000-8000-000000000205','workflow.publication','publish'),
  ('019c0000-0418-7000-8000-000000000201','workflow.instance','start'),
  ('019c0000-5100-7000-8000-000000000206','workflow.instance','read'),
  ('019c0000-5100-7000-8000-000000000207','workflow.instance','signal'),
  ('019c0000-5100-7000-8000-000000000208','workflow.instance','cancel'),
  ('019c0000-5100-7000-8000-000000000209','workflow.instance','replan'),
  ('019c0000-5100-7000-8000-000000000210','workflow.instance','repair'),
  ('019c0000-5100-7000-8000-000000000211','workflow.human_tasks','read'),
  ('019c0000-0500-7000-8000-000000000201','workflow.human_tasks','claim'),
  ('019c0000-0500-7000-8000-000000000202','workflow.human_tasks','decide'),
  ('019c0000-0500-7000-8000-000000000203','workflow.human_tasks','reassign'),
  ('019c0000-0500-7000-8000-000000000204','workflow.human_tasks','expire'),
  ('019c0000-0500-7000-8000-000000000205','workflow.human_tasks','cancel')
 ) AS expected(id,resource_key,action_key)
 JOIN iam_resources AS resource ON resource.resource_key=expected.resource_key
 JOIN iam_actions AS action ON action.action_key=expected.action_key
) AS expected
ON target.match_kind='exact' AND target.resource_id=expected.resource_id
AND target.action_id=expected.action_id
WHEN MATCHED AND (
 target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active' OR
 pg_catalog.substr(target.id::text,20,1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
 id,match_kind,resource_id,action_id,is_protected,state,revision,
 created_at,updated_at,retired_at
) VALUES (
 expected.id,'exact',expected.resource_id,expected.action_id,false,'active',1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp(),NULL
);
