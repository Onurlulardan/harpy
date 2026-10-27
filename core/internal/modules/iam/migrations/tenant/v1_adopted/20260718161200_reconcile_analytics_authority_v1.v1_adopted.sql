-- +goose Up
MERGE INTO iam_resources AS target
USING (VALUES
 ('019c8000-1200-7000-8000-000000000001'::uuid,'analytics.query'::text,'analytics_query'::text,'Analytics Queries'::text),
 ('019c8000-1200-7000-8000-000000000002','analytics.dashboard','analytics_dashboard','Analytics Dashboards'),
 ('019c8000-1200-7000-8000-000000000003','analytics.publication','analytics_publication','Analytics Publications')
) AS expected(id,resource_key,resource_kind,display_name)
ON target.resource_key=expected.resource_key
WHEN MATCHED AND (
 target.owner_domain IS DISTINCT FROM 'analytics' OR
 target.resource_kind IS DISTINCT FROM expected.resource_kind OR
 target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
 id,resource_key,owner_domain,resource_kind,display_name,description,
 is_protected,state,revision,created_at,updated_at,retired_at
) VALUES (
 expected.id,expected.resource_key,'analytics',expected.resource_kind,
 expected.display_name,NULL,false,'active',1,pg_catalog.transaction_timestamp(),
 pg_catalog.transaction_timestamp(),NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
 ('019c8000-1200-7000-8000-000000000101'::uuid,'create'::text,'Create'::text),
 ('019c8000-1200-7000-8000-000000000102','read','Read'),
 ('019c8000-1200-7000-8000-000000000103','update','Update'),
 ('019c8000-1200-7000-8000-000000000104','delete','Delete'),
 ('019c8000-1200-7000-8000-000000000105','publish','Publish'),
 ('019c8000-1200-7000-8000-000000000106','execute','Execute'),
 ('019c8000-1200-7000-8000-000000000107','revoke','Revoke'),
 ('019c8000-1200-7000-8000-000000000108','publish_public','Publish Public'),
 ('019c8000-1200-7000-8000-000000000109','approve_public','Approve Public'),
 ('019c8000-1200-7000-8000-000000000110','publish_capability','Publish Capability'),
 ('019c8000-1200-7000-8000-000000000111','approve_capability','Approve Capability')
) AS expected(id,action_key,display_name)
ON target.action_key=expected.action_key
WHEN MATCHED AND (
 target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active'
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
  ('019c8000-1200-7000-8000-000000000201'::uuid,'analytics.query'::text,'create'::text),
  ('019c8000-1200-7000-8000-000000000202','analytics.query','read'),
  ('019c8000-1200-7000-8000-000000000203','analytics.query','update'),
  ('019c8000-1200-7000-8000-000000000204','analytics.query','delete'),
  ('019c8000-1200-7000-8000-000000000205','analytics.query','publish'),
  ('019c8000-1200-7000-8000-000000000206','analytics.query','execute'),
  ('019c8000-1200-7000-8000-000000000207','analytics.dashboard','create'),
  ('019c8000-1200-7000-8000-000000000208','analytics.dashboard','read'),
  ('019c8000-1200-7000-8000-000000000209','analytics.dashboard','update'),
  ('019c8000-1200-7000-8000-000000000210','analytics.dashboard','delete'),
  ('019c8000-1200-7000-8000-000000000211','analytics.dashboard','publish'),
  ('019c8000-1200-7000-8000-000000000212','analytics.publication','read'),
  ('019c8000-1200-7000-8000-000000000213','analytics.publication','revoke'),
  ('019c8000-1200-7000-8000-000000000214','analytics.publication','publish_public'),
  ('019c8000-1200-7000-8000-000000000215','analytics.publication','approve_public'),
  ('019c8000-1200-7000-8000-000000000216','analytics.publication','publish_capability'),
  ('019c8000-1200-7000-8000-000000000217','analytics.publication','approve_capability')
 ) AS expected(id,resource_key,action_key)
 JOIN iam_resources AS resource ON resource.resource_key=expected.resource_key
 JOIN iam_actions AS action ON action.action_key=expected.action_key
) AS expected
ON target.match_kind='exact' AND target.resource_id=expected.resource_id
AND target.action_id=expected.action_id
WHEN MATCHED AND (
 target.is_protected IS DISTINCT FROM false OR target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
 id,match_kind,resource_id,action_id,is_protected,state,revision,
 created_at,updated_at,retired_at
) VALUES (
 expected.id,'exact',expected.resource_id,expected.action_id,false,'active',1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp(),NULL
);
