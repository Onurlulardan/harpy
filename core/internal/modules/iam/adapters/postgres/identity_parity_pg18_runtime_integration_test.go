//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func applyIAMIdentityParityRuntimeTargets(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents := iamProfileRepairDocuments(t, origin, []string{
		"20260801015110_create_iam_identity_user_reads_v3",
		"20260801015111_create_iam_identity_organization_reads_v2",
		"20260801015120_create_iam_identity_bulk_delete_receipts_v1",
	})
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(
			t, ctx, database, tenant.schema, documents,
		)
	}
}

func assertIAMIdentityParityRuntime(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, "t_iam_a", `
	UPDATE iam_organization_memberships SET membership_kind='owner',
	 revision=revision+1,updated_at=updated_at+interval '1 second'
WHERE id='a4100000-0000-4000-8000-000000000001'`)
	actor := readIAMIdentityActor(
		t, ctx, database, "t_iam_a", iamIdentityActorID,
	)
	withIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(tx *sql.Tx) {
		var verified bool
		var modes, organizations string
		err := tx.QueryRowContext(ctx, `
SELECT email_verified,pg_catalog.array_to_string(access_modes,','),
       pg_catalog.array_to_string(organization_ids,',')
FROM iam_identity_user_read_v3(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5)`,
			iamIdentityReadTargetID, iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion,
		).Scan(&verified, &modes, &organizations)
		if err != nil || !verified || modes != "portal" ||
			organizations != "a2100000-0000-4000-8000-000000000001" {
			t.Fatalf("user v3=%t/%q/%q error=%v",
				verified, modes, organizations, err)
		}
		var owners string
		err = tx.QueryRowContext(ctx, `
SELECT pg_catalog.array_to_string(owner_ids,',')
FROM iam_identity_organization_read_v2(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5)`,
			"a2100000-0000-4000-8000-000000000001", iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion,
		).Scan(&owners)
		if err != nil || owners != iamIdentityReadTargetID {
			t.Fatalf("organization v2 owners=%q error=%v", owners, err)
		}
	})
	assertIAMIdentityParityReadACL(t, ctx, database)
	assertIAMIdentityBulkDeleteReceipt(t, ctx, database)
}

func assertIAMIdentityParityReadACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for signature, want := range map[string]bool{
		"t_iam_a.iam_identity_user_read_v2(name,uuid,uuid,uuid,bigint,bigint)":         false,
		"t_iam_a.iam_identity_user_read_v3(name,uuid,uuid,uuid,bigint,bigint)":         true,
		"t_iam_a.iam_identity_organization_read_v1(name,uuid,uuid,uuid,bigint,bigint)": false,
		"t_iam_a.iam_identity_organization_read_v2(name,uuid,uuid,uuid,bigint,bigint)": true,
	} {
		var allowed bool
		err := database.QueryRowContext(ctx,
			"SELECT pg_catalog.has_function_privilege('harpy_runtime',$1,'EXECUTE')",
			signature).Scan(&allowed)
		if err != nil || allowed != want {
			t.Fatalf("IAM parity ACL %s=%t want=%t error=%v",
				signature, allowed, want, err)
		}
	}
}
