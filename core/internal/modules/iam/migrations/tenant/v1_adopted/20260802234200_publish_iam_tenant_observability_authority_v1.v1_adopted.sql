-- +goose Up
MERGE INTO iam_resources AS target
USING (VALUES
    ('019fb202-0000-8000-8000-000000000001'::uuid,
     'changelog'::text, 'audit'::text, 'audit_changelog'::text,
     'Change Log'::text, 'Tenant data-change history.'::text),
    ('019fb202-0000-8000-8000-000000000002',
     'analytics.usage', 'usage', 'usage_analytics',
     'Usage Analytics', 'Tenant usage analytics.'::text)
) AS expected(id,resource_key,owner_domain,resource_kind,display_name,description)
ON target.resource_key=expected.resource_key
WHEN MATCHED AND (
    target.owner_domain IS DISTINCT FROM expected.owner_domain OR
    target.resource_kind IS DISTINCT FROM expected.resource_kind OR
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
    id,resource_key,owner_domain,resource_kind,display_name,description,
    is_protected,state,revision,created_at,updated_at,retired_at
) VALUES (
    expected.id,expected.resource_key,expected.owner_domain,
    expected.resource_kind,expected.display_name,expected.description,
    true,'active',1,pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(),NULL
);

MERGE INTO iam_actions AS target
USING (VALUES (
    '019fb202-0000-8000-8000-000000000101'::uuid,
    'read'::text, 'Read'::text
)) AS expected(id,action_key,display_name)
ON target.action_key=expected.action_key
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM false OR
    target.state IS DISTINCT FROM 'active'
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
    SELECT expected.id,resource.id AS resource_id,action.id AS action_id
    FROM (VALUES
        ('019fb202-0000-8000-8000-000000000201'::uuid,'changelog'::text),
        ('019fb202-0000-8000-8000-000000000202','analytics.usage')
    ) AS expected(id,resource_key)
    JOIN iam_resources AS resource
      ON resource.resource_key=expected.resource_key
    JOIN iam_actions AS action ON action.action_key='read'
) AS expected
ON target.match_kind='exact'
AND target.resource_id=expected.resource_id
AND target.action_id=expected.action_id
WHEN MATCHED AND (
    target.is_protected IS DISTINCT FROM true OR
    target.state IS DISTINCT FROM 'active'
) THEN UPDATE SET revision=target.revision
WHEN NOT MATCHED THEN INSERT (
    id,match_kind,resource_id,action_id,is_protected,
    state,revision,created_at,updated_at,retired_at
) VALUES (
    expected.id,'exact',expected.resource_id,expected.action_id,true,
    'active',1,pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(),NULL
);

MERGE INTO iam_role_permission_bindings AS target
USING (
    SELECT role.id AS role_id,role.is_protected,
           permission.id AS permission_id
    FROM (VALUES
        ('changelog'::text,'Tenant Yöneticisi'::text),
        ('changelog','Genel Müdür'),
        ('changelog','Temel Kullanıcı'),
        ('changelog','Yönetici'),
        ('changelog','Müdür'),
        ('changelog','Kullanıcı'),
        ('analytics.usage','Tenant Yöneticisi'),
        ('analytics.usage','Yönetici')
    ) AS grant_map(resource_key,role_name)
    JOIN iam_roles AS role ON role.name=grant_map.role_name
    JOIN iam_resources AS resource
      ON resource.resource_key=grant_map.resource_key
    JOIN iam_actions AS action ON action.action_key='read'
    JOIN iam_permissions AS permission
      ON permission.match_kind='exact'
     AND permission.resource_id=resource.id
     AND permission.action_id=action.id
    WHERE role.status='active' AND NOT role.is_protected
      AND resource.state='active' AND action.state='active'
      AND permission.state='active'
) AS expected
ON target.role_id=expected.role_id
AND target.permission_id=expected.permission_id
WHEN MATCHED AND target.state='revoked' THEN UPDATE SET
    state='active',revision=target.revision+1,
    updated_at=pg_catalog.transaction_timestamp(),
    activated_at=pg_catalog.transaction_timestamp(),revoked_at=NULL
WHEN NOT MATCHED THEN INSERT (
    id,role_id,role_is_protected,permission_id,permission_match_kind,
    state,revision,created_at,updated_at,activated_at,revoked_at
) VALUES (
    pg_catalog.uuidv7(),expected.role_id,expected.is_protected,
    expected.permission_id,'exact','active',1,
    pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp(),
    pg_catalog.transaction_timestamp(),NULL
);

WITH bumped AS (
    UPDATE iam_authorization_clock
    SET clock_value=clock_value+1,revision=revision+1,
        updated_at=pg_catalog.transaction_timestamp()
    WHERE singleton
    RETURNING clock_value
)
SELECT iam_rbac_bump_all_users_v1(clock_value) FROM bumped;

SELECT 1/CASE WHEN (
    SELECT pg_catalog.count(*)
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id=permission.resource_id
    JOIN iam_actions AS action ON action.id=permission.action_id
    WHERE permission.match_kind='exact' AND permission.is_protected
      AND permission.state='active' AND action.action_key='read'
      AND action.state='active' AND NOT action.is_protected
      AND resource.is_protected AND resource.state='active'
      AND (resource.resource_key,resource.owner_domain,resource.resource_kind)
          IN (('changelog','audit','audit_changelog'),
              ('analytics.usage','usage','usage_analytics'))
)=2 AND NOT EXISTS (
    SELECT 1
    FROM (VALUES
        ('changelog'::text,'Tenant Yöneticisi'::text),
        ('changelog','Genel Müdür'),('changelog','Temel Kullanıcı'),
        ('changelog','Yönetici'),('changelog','Müdür'),('changelog','Kullanıcı'),
        ('analytics.usage','Tenant Yöneticisi'),('analytics.usage','Yönetici')
    ) AS grant_map(resource_key,role_name)
    JOIN iam_roles AS role ON role.name=grant_map.role_name
    JOIN iam_resources AS resource ON resource.resource_key=grant_map.resource_key
    JOIN iam_permissions AS permission ON permission.resource_id=resource.id
    JOIN iam_actions AS action ON action.id=permission.action_id
    WHERE role.status='active' AND NOT role.is_protected
      AND permission.match_kind='exact' AND permission.state='active'
      AND action.action_key='read' AND NOT EXISTS (
          SELECT 1 FROM iam_role_permission_bindings AS binding
          WHERE binding.role_id=role.id AND binding.permission_id=permission.id
            AND binding.permission_match_kind='exact' AND binding.state='active'
      )
) THEN 1 ELSE 0 END AS iam_tenant_observability_authority_assertion;
