//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	iamRelationCreateUserID    = "c1000000-0000-4000-8000-000000000001"
	iamRelationUpdateUserID    = "c1000000-0000-4000-8000-000000000002"
	iamRelationSecurityUserID  = "c1000000-0000-4000-8000-000000000003"
	iamRelationRollbackUserID  = "c1000000-0000-4000-8000-000000000004"
	iamRelationIsolationUserID = "c1000000-0000-4000-8000-000000000005"
	iamRelationOrgOneID        = "b1000000-0000-4000-8000-000000000001"
	iamRelationOrgTwoID        = "b1000000-0000-4000-8000-000000000002"
	iamRelationOrgRoleID       = "b2000000-0000-4000-8000-000000000001"
	iamRelationDefaultRoleID   = "a1000000-0000-4000-8000-000000000001"
)

const seedIAMRelationUsers = `
INSERT INTO iam_users (
    id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
    profile_values,profile_sha256,token_version,authorization_version,revision,
    created_at,updated_at,tombstoned_at
) SELECT id,'human','active','user',
    '41000000-0000-4000-8000-000000000001','{}'::jsonb,
    pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,0,1,
    '2026-07-16T00:30:00Z','2026-07-16T00:30:00Z',NULL
FROM pg_catalog.unnest(ARRAY[
    'c1000000-0000-4000-8000-000000000001'::uuid,
    'c1000000-0000-4000-8000-000000000002'::uuid,
    'c1000000-0000-4000-8000-000000000003'::uuid,
    'c1000000-0000-4000-8000-000000000004'::uuid,
    'c1000000-0000-4000-8000-000000000005'::uuid
]) AS id`

const seedIAMRelationOrganizations = `
INSERT INTO iam_organizations (
    id,name,slug,status,parent_id,profile_resource_kind,
    profile_schema_version_id,profile_values,profile_sha256,
    revision,created_at,updated_at,tombstoned_at
) VALUES
    ('b1000000-0000-4000-8000-000000000001','Relation One','relation-one',
     'active',NULL,'organization','41000000-0000-4000-8000-000000000002',
     '{}'::jsonb,pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,
     '2026-07-16T00:30:00Z','2026-07-16T00:30:00Z',NULL),
    ('b1000000-0000-4000-8000-000000000002','Relation Two','relation-two',
     'active',NULL,'organization','41000000-0000-4000-8000-000000000002',
     '{}'::jsonb,pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,
     '2026-07-16T00:30:00Z','2026-07-16T00:30:00Z',NULL)`

const seedIAMRelationOrganizationRole = `
INSERT INTO iam_roles (
    id,scope_kind,scope_id,organization_id,name,name_key,description,status,
    is_protected,is_default,profile_resource_kind,profile_schema_version_id,
    profile_values,profile_sha256,revision,created_at,updated_at,retired_at
) VALUES (
    'b2000000-0000-4000-8000-000000000001','organization',
    'b1000000-0000-4000-8000-000000000001',
    'b1000000-0000-4000-8000-000000000001','Organization Operator',
    pg_catalog.convert_to('Organization Operator','UTF8'),NULL,'active',false,false,
    'role','41000000-0000-4000-8000-000000000003','{}'::jsonb,
    pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,
    '2026-07-16T00:30:00Z','2026-07-16T00:30:00Z',NULL)`

const seedIAMRelationOwnerMembership = `
INSERT INTO iam_organization_memberships (
    id,organization_id,user_id,membership_kind,state,revision,
    created_at,updated_at,tombstoned_at
) VALUES (
    'b3000000-0000-4000-8000-000000000001',
    'b1000000-0000-4000-8000-000000000002',
    '51000000-0000-4000-8000-000000000002','owner','active',1,
    '2026-07-16T00:30:00Z','2026-07-16T00:30:00Z',NULL)`

func seedIAMUserRelationsPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		for _, statement := range []string{
			seedIAMRelationUsers,
			seedIAMRelationOrganizations,
			seedIAMRelationOrganizationRole,
			seedIAMRelationOwnerMembership,
		} {
			mustExecIAMTenant(t, ctx, database, tenant.schema, statement)
		}
	}
}
