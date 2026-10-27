//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const restQueryIntegerFieldID = "e1000000-0000-4000-8000-000000000001"
const restQueryTenantFieldID = "e1000000-0000-4000-8000-000000000002"
const restQueryDisplayNameFieldID = "e1000000-0000-4000-8000-000000000003"

func applyIAMRestQueryMigrations(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamRestQueryMigrationNames))
	for _, name := range iamRestQueryMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func seedIAMRestQueryTenants(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for index, tenant := range iamFoundationTenants {
		marker := "tenant-a"
		if index == 1 {
			marker = "tenant-b"
		}
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_profile_schema_versions (
    id,resource_kind,version,state,definition_sha256,activated_at,
    retired_at,created_at,updated_at,revision
) VALUES
('41000000-0000-4000-8000-000000000001','user',1,'active',
 pg_catalog.decode(pg_catalog.repeat('11',32),'hex'),statement_timestamp(),NULL,
 statement_timestamp(),statement_timestamp(),1),
('41000000-0000-4000-8000-000000000002','organization',1,'active',
 pg_catalog.decode(pg_catalog.repeat('12',32),'hex'),statement_timestamp(),NULL,
 statement_timestamp(),statement_timestamp(),1),
('41000000-0000-4000-8000-000000000003','role',1,'active',
 pg_catalog.decode(pg_catalog.repeat('13',32),'hex'),statement_timestamp(),NULL,
 statement_timestamp(),statement_timestamp(),1);`)
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_profile_fields (
    id,resource_kind,field_key,value_kind,cardinality,required,classification,
    normalization_profile,state,created_at,updated_at,retired_at,revision
) VALUES (
    $1,'user','display_name','string','single',true,'public','none_v1',
    'active',statement_timestamp(),statement_timestamp(),NULL,1
);`, restQueryDisplayNameFieldID)
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_users (
    id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
    profile_values,profile_sha256,token_version,authorization_version,
    revision,created_at,updated_at,tombstoned_at
) VALUES
('51000000-0000-4000-8000-000000000001','human','active','user',
 '41000000-0000-4000-8000-000000000001',
 pg_catalog.jsonb_build_object(
     $1::text,'42',$2::text,$3::text,$4::text,$5::text
 ),
 pg_catalog.decode(pg_catalog.repeat('21',32),'hex'),1,1,1,
 statement_timestamp(),statement_timestamp(),NULL),
('51000000-0000-4000-8000-000000000002','human','active','user',
 '41000000-0000-4000-8000-000000000001',
 pg_catalog.jsonb_build_object('display_name',$6::text),
 pg_catalog.decode(pg_catalog.repeat('22',32),'hex'),1,1,1,
 statement_timestamp(),statement_timestamp(),NULL);`,
			restQueryIntegerFieldID, restQueryTenantFieldID, marker,
			restQueryDisplayNameFieldID, "Straße "+marker, "Legacy "+marker)
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_organizations (
    id,name,slug,status,parent_id,profile_resource_kind,
    profile_schema_version_id,profile_values,profile_sha256,revision,
    created_at,updated_at,tombstoned_at
) VALUES (
    '52000000-0000-4000-8000-000000000001',$1,$2,'active',NULL,
    'organization','41000000-0000-4000-8000-000000000002','{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('23',32),'hex'),1,
    statement_timestamp(),statement_timestamp(),NULL
);`, "MÜNCHEN "+marker, marker+"-organization")
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_organization_memberships (
    id,organization_id,user_id,membership_kind,state,revision,
    created_at,updated_at,tombstoned_at
) VALUES
('54000000-0000-4000-8000-000000000001',
 '52000000-0000-4000-8000-000000000001',
 '51000000-0000-4000-8000-000000000001','owner','active',1,
 statement_timestamp(),statement_timestamp(),NULL),
('54000000-0000-4000-8000-000000000002',
 '52000000-0000-4000-8000-000000000001',
 '51000000-0000-4000-8000-000000000002','member','active',1,
 statement_timestamp(),statement_timestamp(),NULL);`)
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO iam_roles (
    id,scope_kind,scope_id,organization_id,name,name_key,description,status,
    is_protected,is_default,profile_resource_kind,profile_schema_version_id,
    profile_values,profile_sha256,revision,created_at,updated_at,retired_at
) VALUES (
    '53000000-0000-4000-8000-000000000001','tenant',$1,NULL,$2,
    pg_catalog.convert_to($2,'UTF8'),'Straße yönetimi','active',false,false,
    'role','41000000-0000-4000-8000-000000000003','{}'::jsonb,
    pg_catalog.decode(pg_catalog.repeat('24',32),'hex'),1,
    statement_timestamp(),statement_timestamp(),NULL
);`, tenant.id, "Straße Operator "+marker)
	}
}
