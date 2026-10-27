//go:build integration

package postgres

const seedIAMTenantObservabilityRoles = `
INSERT INTO iam_roles (
    id,scope_kind,scope_id,organization_id,name,name_key,description,status,
    is_protected,is_default,profile_resource_kind,profile_schema_version_id,
    profile_values,profile_sha256,revision,created_at,updated_at,retired_at
)
SELECT expected.id,'tenant',$1,NULL,expected.name,
       pg_catalog.convert_to(expected.name,'UTF8'),NULL,'active',false,
       expected.is_default,'role','41000000-0000-4000-8000-000000000003',
       '{}'::jsonb,pg_catalog.decode(pg_catalog.repeat(expected.digest,32),'hex'),
       1,'2026-08-02T00:00:00Z','2026-08-02T00:00:00Z',NULL
FROM (VALUES
 ('97100000-0000-4000-8000-000000000001'::uuid,'Temel Kullanıcı'::text,true,'81'::text),
 ('97100000-0000-4000-8000-000000000002','Tenant Yöneticisi',false,'82'),
 ('97100000-0000-4000-8000-000000000003','Genel Müdür',false,'83'),
 ('97100000-0000-4000-8000-000000000004','Kullanıcı',false,'84'),
 ('97100000-0000-4000-8000-000000000005','Yönetici',false,'85'),
 ('97100000-0000-4000-8000-000000000006','Müdür',false,'86'),
 ('97100000-0000-4000-8000-000000000007','BT Uzmanı',false,'87')
) AS expected(id,name,is_default,digest)
`
