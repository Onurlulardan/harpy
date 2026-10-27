-- +goose Up
-- Notification HTTP authority is exact and idempotent; semantic drift aborts.
MERGE INTO iam_resources AS target
USING (VALUES
 ('019d0000-2203-7000-8000-000000000001'::uuid,'notifications.inbox'::text,'notification_inbox'::text,'Notification Inbox'::text),
 ('019d0000-2203-7000-8000-000000000002','notifications.preferences','notification_preferences','Notification Preferences'),
 ('019d0000-2203-7000-8000-000000000003','notifications.intent','notification_intent','Notification Intents'),
 ('019d0000-2203-7000-8000-000000000004','notifications.realtime','notification_realtime','Notification Realtime')
) AS expected(id,resource_key,resource_kind,display_name)
ON target.resource_key=expected.resource_key
WHEN MATCHED AND (
 target.owner_domain IS DISTINCT FROM 'notifications' OR
 target.resource_kind IS DISTINCT FROM expected.resource_kind OR
 target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active' OR
 pg_catalog.substr(target.id::text,20,1) NOT IN ('8','9','a','b')
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
 id,resource_key,owner_domain,resource_kind,display_name,description,
 is_protected,state,revision,created_at,updated_at,retired_at
) VALUES (
 expected.id,expected.resource_key,'notifications',expected.resource_kind,
 expected.display_name,NULL,false,'active',1,pg_catalog.transaction_timestamp(),
 pg_catalog.transaction_timestamp(),NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
 ('019d0000-2203-7000-8000-000000000101'::uuid,'read'::text,'Read'::text),
 ('019d0000-2203-7000-8000-000000000102','update','Update'),
 ('019d0000-2203-7000-8000-000000000103','delete','Delete'),
 ('019d0000-2203-7000-8000-000000000104','send','Send'),
 ('019d0000-2203-7000-8000-000000000105','connect','Connect')
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
  ('019d0000-2203-7000-8000-000000000201'::uuid,'notifications.inbox'::text,'read'::text),
  ('019d0000-2203-7000-8000-000000000202','notifications.inbox','update'),
  ('019d0000-2203-7000-8000-000000000203','notifications.inbox','delete'),
  ('019d0000-2203-7000-8000-000000000204','notifications.preferences','read'),
  ('019d0000-2203-7000-8000-000000000205','notifications.preferences','update'),
  ('019d0000-2203-7000-8000-000000000206','notifications.intent','send'),
  ('019d0000-2203-7000-8000-000000000207','notifications.realtime','connect')
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
