//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestIAMLifecycleRuntimeAppliesAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "lifecycle_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			prepareIAMIdentityRuntimeAuthority(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMIdentityPermissionResolver(t, ctx, database, origin)
			applyIAMLifecycleIdentityAuthority(t, ctx, database, origin)
			applyIAMLifecycleRuntime(t, ctx, database, origin)
			assertIAMLifecycleRuntimeObjects(t, ctx, database)
			seedIAMFoundationBase(t, ctx, database)
			seedIAMIdentityRuntime(t, ctx, database)
			seedIAMLifecycleRuntimeSessions(t, ctx, database)
			assertIAMLifecycleRuntimeBehavior(t, ctx, database)
		})
	}
}

func applyIAMLifecycleIdentityAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	document := iamFoundationDocument{
		path: iamIdentityTenantMigrationNames[0] + "." + origin + ".sql",
		content: readIAMIdentityMigration(
			t, origin, iamIdentityTenantMigrationNames[0],
		),
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, []iamFoundationDocument{document})
	}
}

func applyIAMLifecycleRuntime(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	names := append(
		append([]string(nil), iamLifecycleMigrationNames...),
		iamLifecycleRuntimeMigrationNames...,
	)
	names = append(names, iamLifecycleStartV2QualificationMigrations...)
	names = append(names, iamLifecycleKeyContractMigrations...)
	documents := make([]iamFoundationDocument, 0, len(names))
	for _, name := range names {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func assertIAMLifecycleRuntimeObjects(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `
SELECT
    pg_catalog.to_regclass($1) IS NOT NULL,
    pg_catalog.to_regclass($2) IS NOT NULL,
    pg_catalog.count(DISTINCT procedure.proname)
FROM pg_catalog.pg_proc AS procedure
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = procedure.pronamespace
WHERE namespace.nspname = $3
  AND procedure.proname = ANY (ARRAY[
    'iam_lifecycle_replay_v1','iam_lifecycle_record_v1',
    'iam_identifier_add_v1','iam_identifier_verify_v1',
    'iam_identifier_tombstone_v1','iam_identifier_retire_v1',
    'iam_one_time_issue_v1','iam_email_verification_start_v1',
    'iam_password_reset_start_v1','iam_portal_setup_start_v1',
    'iam_password_reset_complete_v1','iam_portal_setup_complete_v2',
    'iam_portal_approval_decide_v1','iam_password_change_v1',
    'iam_authenticator_list_v1','iam_authenticator_suspend_v1',
    'iam_authenticator_retire_v1','iam_webauthn_enrollment_start_v1',
    'iam_webauthn_enrollment_complete_v1'
  ])`
	for _, tenant := range iamFoundationTenants {
		var receipts, audit bool
		var functions int
		err := database.QueryRowContext(
			ctx, statement, tenant.schema+".iam_lifecycle_operation_receipts",
			tenant.schema+".iam_lifecycle_audit", tenant.schema,
		).Scan(&receipts, &audit, &functions)
		if err != nil {
			t.Fatalf("inspect %s IAM lifecycle runtime: %v", tenant.schema, err)
		}
		if !receipts || !audit || functions != 19 {
			t.Fatalf(
				"%s lifecycle objects = receipts:%t audit:%t functions:%d",
				tenant.schema, receipts, audit, functions,
			)
		}
	}
	assertIAMLifecycleV2RuntimeObjects(t, ctx, database)
}

func assertIAMLifecycleV2RuntimeObjects(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const statement = `
SELECT
    pg_catalog.to_regclass($1) IS NOT NULL,
    pg_catalog.to_regclass($2) IS NOT NULL,
    pg_catalog.to_regclass($3) IS NOT NULL,
    pg_catalog.to_regprocedure($4) IS NOT NULL,
    pg_catalog.to_regprocedure($5) IS NOT NULL,
    pg_catalog.to_regprocedure($6) IS NOT NULL`
	for _, tenant := range iamFoundationTenants {
		var artifacts, claims, receipts, identifier, totp, sessions bool
		err := database.QueryRowContext(ctx, statement,
			tenant.schema+".iam_lifecycle_delivery_artifacts",
			tenant.schema+".iam_lifecycle_delivery_worker_claims",
			tenant.schema+".iam_lifecycle_delivery_worker_receipts",
			tenant.schema+".iam_identifier_verify_v2(name,uuid,uuid,bytea,bytea,text,bytea,bytea)",
			tenant.schema+".iam_totp_enrollment_complete_v1(name,uuid,uuid,bytea,uuid,uuid,text,bytea,bytea,bigint,bigint,bigint,uuid,uuid,bigint,bigint)",
			tenant.schema+".iam_session_revoke_all_v1(name,uuid,uuid,bytea,uuid,uuid,bigint,bigint)",
		).Scan(&artifacts, &claims, &receipts, &identifier, &totp, &sessions)
		if err != nil || !artifacts || !claims || !receipts || !identifier || !totp || !sessions {
			t.Fatalf("%s lifecycle V2 objects=%t/%t/%t/%t/%t/%t err=%v",
				tenant.schema, artifacts, claims, receipts, identifier, totp, sessions, err)
		}
	}
}
