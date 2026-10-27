//go:build integration

package postgres

const iamIdentityDeleteOwnedOrganizationCreateSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_organization_create_v3(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,NULL,ARRAY[$6::uuid],ARRAY[$7::uuid],
 $8,$9,$10,$11,$12,$13,$14)`

const iamIdentityDeleteRoleCreateSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_role_create_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`

const iamIdentityDeleteUserCreateSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_user_create_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`

const iamIdentityDeleteMembershipAddSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_membership_add_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`

const iamIdentityDeleteAssignmentGrantSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_assignment_grant_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

const iamIdentityDeleteUserDisableSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_user_disable_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8)`

const iamIdentityDeleteAssignmentRevokeSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_assignment_revoke_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8)`

const iamIdentityDeleteMembershipDeleteSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_membership_delete_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8)`

const iamIdentityDeleteUserDeleteSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_user_delete_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8)`

const iamIdentityDeleteRoleRetireSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_role_retire_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8)`

const iamIdentityDeleteOrganizationDisableSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_organization_disable_v1(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8)`

const iamIdentityDeleteOrganizationDeleteSQL = `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_organization_delete_v2(
 (pg_catalog.current_schemas(false))[2]::name,
 $1,$2,$3,$4,$5,$6,$7,$8)`
