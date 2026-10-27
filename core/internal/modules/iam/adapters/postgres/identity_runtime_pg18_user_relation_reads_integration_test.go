//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const iamIdentityReadTargetID = "51000000-0000-4000-8000-000000000002"

const seedIAMIdentityReadRelations = `
INSERT INTO iam_user_identifiers (
    id,user_id,kind,display_value,normalized_value,normalized_sha256,state,
    revision,verified_at,activated_at,tombstoned_at,cooldown_until,retired_at,
    created_at,updated_at
) VALUES (
    '61000000-0000-4000-8000-000000000002',
    '51000000-0000-4000-8000-000000000002','email_v1',
    'Target@example.com','target@example.com',
    pg_catalog.decode(pg_catalog.repeat('35',32),'hex'),'active',1,
    '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL,NULL,NULL,
    '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z'
);
INSERT INTO iam_organizations (
    id,name,slug,status,parent_id,profile_resource_kind,
    profile_schema_version_id,profile_values,profile_sha256,
    revision,created_at,updated_at,tombstoned_at
) VALUES
    ('a2100000-0000-4000-8000-000000000001','Active Org','active-org','active',
     NULL,'organization','41000000-0000-4000-8000-000000000002','{}'::jsonb,
     pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL),
    ('a2200000-0000-4000-8000-000000000001','Suspended Org','suspended-org','active',
     NULL,'organization','41000000-0000-4000-8000-000000000002','{}'::jsonb,
     pg_catalog.sha256(pg_catalog.convert_to('{}','UTF8')),1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL);
INSERT INTO iam_organization_memberships (
    id,organization_id,user_id,membership_kind,state,revision,
    created_at,updated_at,tombstoned_at
) VALUES
    ('a4100000-0000-4000-8000-000000000001',
     'a2100000-0000-4000-8000-000000000001',
     '51000000-0000-4000-8000-000000000002','member','active',1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL),
    ('a4200000-0000-4000-8000-000000000001',
     'a2200000-0000-4000-8000-000000000001',
     '51000000-0000-4000-8000-000000000002','member','suspended',1,
     '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL);
INSERT INTO iam_user_role_assignments (
    id,user_id,role_id,role_scope_kind,role_scope_id,
    organization_id,membership_id,state,revision,
    created_at,updated_at,revoked_at
) VALUES (
    'a5100000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000002',
    'a1000000-0000-4000-8000-000000000001','tenant',
    '10000000-0000-4000-8000-000000000001',
    NULL,NULL,'suspended',1,
    '2026-07-16T00:20:00Z','2026-07-16T00:20:00Z',NULL
)`

func assertIAMIdentityUserReadsActiveRelations(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, "t_iam_a", seedIAMIdentityReadRelations)
	assertIAMIdentityUserReadACL(t, ctx, database)
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	withIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(transaction *sql.Tx) {
		assertIAMIdentityRelationProjection(t, transaction.QueryRowContext(ctx, `
SELECT pg_catalog.array_to_string(role_ids,','),
       pg_catalog.array_to_string(organization_ids,',')
FROM iam_identity_user_read_v2(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5)`,
			iamIdentityReadTargetID, iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion))
		assertIAMIdentityRelationProjection(t, transaction.QueryRowContext(ctx, `
SELECT pg_catalog.array_to_string(role_ids,','),
       pg_catalog.array_to_string(organization_ids,',')
FROM iam_identity_user_list_v2(
    (pg_catalog.current_schemas(false))[2]::name,NULL,100,$2,$3,$4,$5)
WHERE id=$1`, iamIdentityReadTargetID, iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion))
	})
}

func assertIAMIdentityRelationProjection(t *testing.T, row *sql.Row) {
	t.Helper()
	var roles, organizations string
	if err := row.Scan(&roles, &organizations); err != nil {
		t.Fatalf("read IAM active user relations: %v", err)
	}
	if roles != "94000000-0000-4000-8000-000000000001" ||
		organizations != "a2100000-0000-4000-8000-000000000001" {
		t.Fatalf("IAM relation projection includes inactive rows: %q / %q",
			roles, organizations)
	}
}

func assertIAMIdentityUserReadACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for signature, want := range map[string]bool{
		"t_iam_a.iam_identity_user_read_v1(name,uuid,uuid,uuid,bigint,bigint)":         false,
		"t_iam_a.iam_identity_user_list_v1(name,uuid,integer,uuid,uuid,bigint,bigint)": false,
		"t_iam_a.iam_identity_user_read_v2(name,uuid,uuid,uuid,bigint,bigint)":         true,
		"t_iam_a.iam_identity_user_list_v2(name,uuid,integer,uuid,uuid,bigint,bigint)": true,
	} {
		var allowed bool
		if err := database.QueryRowContext(ctx,
			"SELECT pg_catalog.has_function_privilege('harpy_runtime',$1,'EXECUTE')",
			signature).Scan(&allowed); err != nil || allowed != want {
			t.Fatalf("IAM user read ACL %s = %t, want %t: %v",
				signature, allowed, want, err)
		}
	}
}
