//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

const (
	publicDeleteUserID       = "62000000-0000-4000-8000-000000000001"
	publicDeleteIdentifierID = "62000000-0000-4000-8000-000000000002"
	publicDeleteAccessID     = "62000000-0000-4000-8000-000000000003"
	publicDeleteOrgID        = "62000000-0000-4000-8000-000000000004"
	publicDeleteRoleID       = "62000000-0000-4000-8000-000000000005"
	publicDeleteMembershipID = "62000000-0000-4000-8000-000000000006"
	publicDeleteAssignmentID = "62000000-0000-4000-8000-000000000007"
	publicDeleteRoleSchemaID = "61000000-0000-4000-8000-000000000002"
	publicDeleteReadPermID   = "93000000-0000-4000-8000-000000000002"
	publicDeleteDeletePermID = "93000000-0000-4000-8000-000000000003"
)

func seedPublicationPublicDeleteGraph(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	const sql = `
INSERT INTO iam_profile_schema_versions(
 id,resource_kind,version,state,definition_sha256,activated_at,
 created_at,updated_at,revision)
VALUES($1,'role',1,'active',
 pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),
 pg_catalog.now(),pg_catalog.now(),pg_catalog.now(),1);
INSERT INTO iam_users(
 id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,token_version,authorization_version,revision,
 created_at,updated_at,tombstoned_at) VALUES
($2,'human','disabled','user',$3,'{}',
 pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,0,1,
 pg_catalog.now(),pg_catalog.now(),NULL);
INSERT INTO iam_user_identifiers(
 id,user_id,kind,display_value,normalized_value,normalized_sha256,state,revision,
 verified_at,activated_at,created_at,updated_at)
VALUES($4,$2,'email_v1','delete-public@example.invalid',
 'delete-public@example.invalid',
 pg_catalog.sha256(pg_catalog.convert_to('delete-public@example.invalid','UTF8')),
 'active',1,pg_catalog.now(),pg_catalog.now(),pg_catalog.now(),pg_catalog.now());
INSERT INTO iam_user_access_modes(
 id,user_id,mode,state,revision,created_at,updated_at,activated_at)
VALUES($5,$2,'operator_console','active',1,pg_catalog.now(),pg_catalog.now(),pg_catalog.now());
INSERT INTO iam_organizations(
 id,name,slug,status,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,revision,created_at,updated_at)
VALUES($6,'Delete Public','delete-public','active','organization',
 (SELECT id FROM iam_profile_schema_versions
  WHERE resource_kind='organization' AND state='active'),'{}',
 pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,pg_catalog.now(),pg_catalog.now());
INSERT INTO iam_roles(
 id,scope_kind,scope_id,organization_id,name,name_key,status,is_protected,is_default,
 profile_resource_kind,profile_schema_version_id,profile_values,profile_sha256,
 revision,created_at,updated_at)
VALUES($7,'organization',$6,$6,'Delete Member',
 pg_catalog.convert_to('Delete Member','UTF8'),'active',false,false,'role',
 (SELECT id FROM iam_profile_schema_versions
  WHERE resource_kind='role' AND state='active'),'{}',
 pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,pg_catalog.now(),pg_catalog.now());
INSERT INTO iam_organization_memberships(
 id,organization_id,user_id,membership_kind,state,revision,created_at,updated_at)
VALUES($8,$6,$2,'member','active',1,pg_catalog.now(),pg_catalog.now());
INSERT INTO iam_user_role_assignments(
 id,user_id,role_id,role_scope_kind,role_scope_id,organization_id,membership_id,
 state,revision,created_at,updated_at)
VALUES($9,$2,$7,'organization',$6,$6,$8,'active',1,pg_catalog.now(),pg_catalog.now());`
	execPublicationTenant(t, ctx, db, sql,
		publicDeleteRoleSchemaID, publicDeleteUserID, publicationSourceID,
		publicDeleteIdentifierID, publicDeleteAccessID, publicDeleteOrgID,
		publicDeleteRoleID, publicDeleteMembershipID, publicDeleteAssignmentID)
}
