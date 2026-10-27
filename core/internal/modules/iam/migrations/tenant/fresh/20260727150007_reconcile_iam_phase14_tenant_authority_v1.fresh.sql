-- +goose Up
MERGE INTO iam_resources AS target
USING (VALUES
    ('019f9d00-0000-8000-8000-000000000001'::uuid,'catalog'::text,'catalog'::text,'catalog','Catalog'),
    ('019f9d00-0000-8000-8000-000000000002','catalog.menu','catalog','catalog_menu','Catalog Menu'),
    ('019f9d00-0000-8000-8000-000000000003','billing','billing_subscription','Billing'),
    ('019f9d00-0000-8000-8000-000000000004','billing.usage','billing','billing_usage','Billing Usage'),
    ('019f9d00-0000-8000-8000-000000000005','billing.entitlements','billing','billing_entitlements','Billing Entitlements'),
    ('019f9d00-0000-8000-8000-000000000006','documents.folder','documents','document_folder','Document Folders'),
    ('019f9d00-0000-8000-8000-000000000007','documents.item','documents','document_item','Document Items')
) AS expected(id,resource_key,owner_domain,resource_kind,display_name)
ON target.resource_key=expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM expected.owner_domain OR
    target.resource_kind IS DISTINCT FROM expected.resource_kind OR
    target.is_protected OR target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
    id,resource_key,owner_domain,resource_kind,display_name,description,
    is_protected,state,revision,created_at,updated_at,retired_at
) VALUES (
    expected.id,expected.resource_key,expected.owner_domain,expected.resource_kind,
    expected.display_name,NULL,false,'active',1,transaction_timestamp(),transaction_timestamp(),NULL
);

MERGE INTO iam_actions AS target
USING (VALUES
    ('019f9d00-0000-8000-8000-000000000101'::uuid,'install'::text,'Install'::text),
    ('019f9d00-0000-8000-8000-000000000102','upgrade','Upgrade'),
    ('019f9d00-0000-8000-8000-000000000103','disable','Disable'),
    ('019f9d00-0000-8000-8000-000000000104','enable','Enable'),
    ('019f9d00-0000-8000-8000-000000000105','restore','Restore'),
    ('019f9d00-0000-8000-8000-000000000106','purge','Purge'),
    ('019f9d00-0000-8000-8000-000000000107','move','Move'),
    ('019f9d00-0000-8000-8000-000000000108','copy','Copy'),
    ('019f9d00-0000-8000-8000-000000000109','download','Download'),
    ('019f9d00-0000-8000-8000-000000000110','version','Version')
) AS expected(id,action_key,display_name)
ON target.action_key=expected.action_key
WHEN MATCHED AND (target.is_protected OR target.state IS DISTINCT FROM 'active')
THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
    id,action_key,display_name,description,is_protected,state,revision,
    created_at,updated_at,retired_at
) VALUES (
    expected.id,expected.action_key,expected.display_name,NULL,false,'active',1,
    transaction_timestamp(),transaction_timestamp(),NULL
);

MERGE INTO iam_permissions AS target
USING (
    SELECT expected.id,resource.id AS resource_id,action.id AS action_id
    FROM (VALUES
        ('019f9d00-0000-8000-8000-000000000201'::uuid,'catalog'::text,'read'::text),
        ('019f9d00-0000-8000-8000-000000000202','catalog','install'),
        ('019f9d00-0000-8000-8000-000000000203','catalog','upgrade'),
        ('019f9d00-0000-8000-8000-000000000204','catalog','disable'),
        ('019f9d00-0000-8000-8000-000000000205','catalog','enable'),
        ('019f9d00-0000-8000-8000-000000000206','catalog','restore'),
        ('019f9d00-0000-8000-8000-000000000207','catalog','retire'),
        ('019f9d00-0000-8000-8000-000000000208','catalog','purge'),
        ('019f9d00-0000-8000-8000-000000000209','catalog.menu','manage'),
        ('019f9d00-0000-8000-8000-000000000210','billing','read'),
        ('019f9d00-0000-8000-8000-000000000211','billing','manage'),
        ('019f9d00-0000-8000-8000-000000000212','billing.usage','read'),
        ('019f9d00-0000-8000-8000-000000000213','billing.entitlements','read'),
        ('019f9d00-0000-8000-8000-000000000214','documents.folder','create'),
        ('019f9d00-0000-8000-8000-000000000215','documents.folder','read'),
        ('019f9d00-0000-8000-8000-000000000216','documents.folder','update'),
        ('019f9d00-0000-8000-8000-000000000217','documents.folder','move'),
        ('019f9d00-0000-8000-8000-000000000218','documents.folder','delete'),
        ('019f9d00-0000-8000-8000-000000000219','documents.item','read'),
        ('019f9d00-0000-8000-8000-000000000220','documents.item','create'),
        ('019f9d00-0000-8000-8000-000000000221','documents.item','update'),
        ('019f9d00-0000-8000-8000-000000000222','documents.item','move'),
        ('019f9d00-0000-8000-8000-000000000223','documents.item','copy'),
        ('019f9d00-0000-8000-8000-000000000224','documents.item','delete'),
        ('019f9d00-0000-8000-8000-000000000225','documents.item','download'),
        ('019f9d00-0000-8000-8000-000000000226','documents.item','version'),
        ('019f9d00-0000-8000-8000-000000000227','documents.item','purge')
    ) AS expected(id,resource_key,action_key)
    JOIN iam_resources AS resource ON resource.resource_key=expected.resource_key
    JOIN iam_actions AS action ON action.action_key=expected.action_key
) AS expected
ON target.match_kind='exact' AND target.resource_id=expected.resource_id
AND target.action_id=expected.action_id
WHEN MATCHED AND (target.is_protected OR target.state IS DISTINCT FROM 'active')
THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
    id,match_kind,resource_id,action_id,is_protected,state,revision,
    created_at,updated_at,retired_at
) VALUES (
    expected.id,'exact',expected.resource_id,expected.action_id,false,'active',1,
    transaction_timestamp(),transaction_timestamp(),NULL
);

SELECT 1/CASE WHEN (
    SELECT count(*) FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id=permission.resource_id
    WHERE permission.match_kind='exact' AND permission.state='active'
      AND resource.resource_key IN (
        'catalog','catalog.menu','billing','billing.usage','billing.entitlements',
        'documents.folder','documents.item'
      )
)=27 THEN 1 ELSE 0 END AS iam_phase14_tenant_authority_assertion;
