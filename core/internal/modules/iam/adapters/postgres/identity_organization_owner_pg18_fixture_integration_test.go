//go:build integration

package postgres

const iamOrganizationOwnerGraphFixture = `
INSERT INTO iam_organization_memberships (
 id,organization_id,user_id,membership_kind,state,revision,created_at,updated_at
) VALUES (
 'b2000000-0000-4000-8000-000000000002',
 'b2000000-0000-4000-8000-000000000001',
 '51000000-0000-4000-8000-000000000002','member','active',1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp()
);
INSERT INTO iam_roles (
 id,scope_kind,scope_id,organization_id,name,name_key,description,status,
 is_protected,is_default,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,revision,created_at,updated_at
) VALUES (
 'b2000000-0000-4000-8000-000000000003','organization',
 'b2000000-0000-4000-8000-000000000001',
 'b2000000-0000-4000-8000-000000000001',
 'Owner Guard',pg_catalog.convert_to('Owner Guard','UTF8'),NULL,'active',
 true,false,'role','41000000-0000-4000-8000-000000000003','{}'::jsonb,
 pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp()
);
INSERT INTO iam_user_role_assignments (
 id,user_id,role_id,role_scope_kind,role_scope_id,organization_id,
 membership_id,state,revision,created_at,updated_at
) VALUES (
 'b2000000-0000-4000-8000-000000000004',
 '51000000-0000-4000-8000-000000000003',
 'b2000000-0000-4000-8000-000000000003','organization',
 'b2000000-0000-4000-8000-000000000001',
 'b2000000-0000-4000-8000-000000000001',
 'b2000000-0000-4000-8000-000000000001',
 'active',1,pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp()
)`

const iamCrossTenantOwnerFixture = `
INSERT INTO iam_users (
 id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,token_version,authorization_version,revision,
 created_at,updated_at,tombstoned_at
) SELECT '51000000-0000-4000-8000-000000000099',identity_kind,status,
 profile_resource_kind,profile_schema_version_id,profile_values,profile_sha256,
 token_version,authorization_version,revision,created_at,updated_at,tombstoned_at
FROM iam_users WHERE id='51000000-0000-4000-8000-000000000003'`

const iamOrganizationOwnerUpdateSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_organization_update_v3(
 (pg_catalog.current_schemas(false))[2]::name,
 $1::uuid,$2::bytea,$3::uuid,$4::bigint,$5::text,$6::text,NULL,
 true,$7::uuid[],$8::uuid[],$9::uuid,$10::bytea,$11::bytea,
 $12::uuid,$13::uuid,$14::bigint,$15::bigint)`

const iamOrganizationNoOwnerCreateSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_organization_create_v3(
 (pg_catalog.current_schemas(false))[2]::name,
 $1::uuid,$2::bytea,$3::uuid,$4::text,$5::text,NULL,
 ARRAY[]::uuid[],ARRAY[]::uuid[],$6::uuid,$7::bytea,$8::bytea,
 $9::uuid,$10::uuid,$11::bigint,$12::bigint)`
