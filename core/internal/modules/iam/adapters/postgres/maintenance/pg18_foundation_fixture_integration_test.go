//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
)

const (
	maintenanceUserSchemaID = "41000000-0000-4000-8000-000000000001"
	maintenanceRoleSchemaID = "41000000-0000-4000-8000-000000000002"
	maintenanceDefaultRole  = "61000000-0000-4000-8000-000000000001"
	maintenanceProfileField = "43000000-0000-4000-8000-000000000010"
	maintenanceUserOne      = "51000000-0000-4000-8000-000000000001"
	maintenanceUserTwo      = "51000000-0000-4000-8000-000000000002"
	maintenanceUserThree    = "51000000-0000-4000-8000-000000000003"
)

var maintenanceProfileUsers = []string{
	maintenanceUserOne, maintenanceUserTwo, maintenanceUserThree,
}

const maintenanceProfileRootSQL = `
INSERT INTO iam_profile_schema_versions(
 id,resource_kind,version,state,definition_sha256,activated_at,retired_at,
 created_at,updated_at,revision)
VALUES('41000000-0000-4000-8000-000000000001','user',1,'active',
 pg_catalog.decode(pg_catalog.repeat('11',32),'hex'),
 '2026-07-16T00:00:00Z',NULL,'2026-07-16T00:00:00Z',
 '2026-07-16T00:00:00Z',1);`

const maintenanceRoleSchemaSQL = `
INSERT INTO iam_profile_schema_versions(
 id,resource_kind,version,state,definition_sha256,activated_at,retired_at,
 created_at,updated_at,revision)
VALUES('41000000-0000-4000-8000-000000000002','role',1,'active',
 pg_catalog.decode(pg_catalog.repeat('12',32),'hex'),
 '2026-07-16T00:00:00Z',NULL,'2026-07-16T00:00:00Z',
 '2026-07-16T00:00:00Z',1);`

const maintenanceDefaultRoleSQL = `
INSERT INTO iam_roles(
 id,scope_kind,scope_id,organization_id,name,name_key,description,status,
 is_protected,is_default,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,revision,created_at,updated_at,retired_at)
VALUES('61000000-0000-4000-8000-000000000001','tenant',$1,NULL,
 'Tenant Member',pg_catalog.convert_to('Tenant Member','UTF8'),NULL,'active',
 false,true,'role','41000000-0000-4000-8000-000000000002','{}'::jsonb,
 pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,
 '2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL);`

const maintenanceUsersSQL = `
INSERT INTO iam_users(
 id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,token_version,authorization_version,revision,
 created_at,updated_at,tombstoned_at)
VALUES
 ('51000000-0000-4000-8000-000000000001','human','active','user',
  '41000000-0000-4000-8000-000000000001','{}'::jsonb,
  pg_catalog.decode(pg_catalog.repeat('21',32),'hex'),1,0,1,
  '2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL),
 ('51000000-0000-4000-8000-000000000002','human','active','user',
  '41000000-0000-4000-8000-000000000001','{}'::jsonb,
  pg_catalog.decode(pg_catalog.repeat('22',32),'hex'),1,0,1,
  '2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL),
 ('51000000-0000-4000-8000-000000000003','human','active','user',
  '41000000-0000-4000-8000-000000000001','{}'::jsonb,
  pg_catalog.decode(pg_catalog.repeat('23',32),'hex'),1,0,1,
  '2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL);`

const maintenanceProfileFieldSQL = `
INSERT INTO iam_profile_fields(
 id,resource_kind,field_key,value_kind,cardinality,required,classification,
 normalization_profile,state,created_at,updated_at,retired_at,revision)
VALUES($1,'user','display_name','string','single',false,'internal','none_v1',
 'active','2026-07-16T01:00:00Z','2026-07-16T01:00:00Z',NULL,1);`

const maintenanceProfileRevisionSQL = `
INSERT INTO iam_profile_field_revisions(
 schema_version_id,resource_kind,field_id,validation_schema,
 validation_sha256,created_at)
VALUES($2,'user',$1,pg_catalog.jsonb_build_object('maximum_bytes',$3::integer),
 pg_catalog.sha256(pg_catalog.convert_to(
  '{"maximum_bytes":'||$3::text||'}','UTF8')),'2026-07-16T01:00:00Z');`

func seedMaintenancePG18Foundation(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for index, tenant := range maintenancePG18Tenants {
		execMaintenancePG18Tenant(t, ctx, database, tenant, maintenanceProfileRootSQL)
		execMaintenancePG18Tenant(t, ctx, database, tenant, maintenanceRoleSchemaSQL)
		execMaintenancePG18Tenant(t, ctx, database, tenant,
			maintenanceDefaultRoleSQL, tenant.id)
		execMaintenancePG18Tenant(t, ctx, database, tenant, maintenanceUsersSQL)
		execMaintenancePG18Tenant(t, ctx, database, tenant,
			maintenanceProfileFieldSQL, maintenanceProfileField)
		execMaintenancePG18Tenant(t, ctx, database, tenant,
			maintenanceProfileRevisionSQL, maintenanceProfileField,
			maintenanceUserSchemaID, 64+index*64)
	}
}
