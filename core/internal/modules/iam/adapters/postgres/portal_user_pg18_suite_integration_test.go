//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

var iamPortalUserMigrationNames = []string{
	"20260723130000_create_iam_portal_user_storage",
	"20260723130001_create_iam_portal_user_create_v1",
	"20260723130002_create_iam_portal_user_status_v1",
	"20260723130003_create_iam_portal_user_console_grant_v1",
	"20260723130004_create_iam_portal_user_console_revoke_v1",
	"20260723130005_create_iam_portal_access_reads_v1",
	"20260723130006_grant_iam_portal_user_runtime_v1",
}

func TestIAMPortalUserProjectionAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "portal_user_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			prepareIAMIdentityRuntimeAuthority(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMIdentityPermissionResolver(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			applyIAMProfileRevisionMigrations(t, ctx, database, origin)
			applyIAMIdentityRuntime(t, ctx, database, origin)
			applyIAMPortalUserMigrations(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			seedIAMPortalProjection(t, ctx, database)
			assertIAMPortalProjection(t, ctx, database)
		})
	}
}

func applyIAMPortalUserMigrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamPortalUserMigrationNames))
	for _, name := range iamPortalUserMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func seedIAMPortalProjection(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, "t_iam_a", `CREATE TABLE tenancy_schema_identity(
 singleton boolean PRIMARY KEY CHECK(singleton),tenant_id uuid NOT NULL,
 schema_name text NOT NULL,origin text NOT NULL,created_at timestamptz NOT NULL)`)
	mustExecIAMTenant(t, ctx, database, "t_iam_a",
		`INSERT INTO tenancy_schema_identity VALUES(true,$1,'t_iam_a','fresh',clock_timestamp())`,
		iamFoundationTenants[0].id)
	mustExecIAMTenant(t, ctx, database, "t_iam_a", `
INSERT INTO iam_authorization_clock VALUES(true,1,1,statement_timestamp(),statement_timestamp());
UPDATE iam_users SET authorization_version=1,revision=revision+1,
	updated_at=statement_timestamp() WHERE id='51000000-0000-4000-8000-000000000001'`)
	mustExecIAMTenant(t, ctx, database, "t_iam_a", `
INSERT INTO iam_portal_subject_bindings(
 id,user_id,access_mode_id,access_mode,subject_shape,subject_type,profile_owner,
 profile_form_id,profile_form_version_id,subject_record_id,identifier_field_id,
 contract_sha256,state,revision,created_at,updated_at,activated_at)
VALUES('018f47f4-7b1b-7000-8000-000000000020',
 '51000000-0000-4000-8000-000000000002','52000000-0000-4000-8000-000000000002',
 'portal','form_record','employee','forms','018f47f4-7b1b-7000-8000-000000000021',
 '018f47f4-7b1b-7000-8000-000000000022','018f47f4-7b1b-7000-8000-000000000023',
 '018f47f4-7b1b-7000-8000-000000000024',decode(repeat('44',32),'hex'),'active',1,
	statement_timestamp(),statement_timestamp(),statement_timestamp())`)
	mustExecIAMTenant(t, ctx, database, "t_iam_a", `
INSERT INTO iam_portal_access_setups(
 binding_id,user_id,access_mode_id,access_mode,setup_state,approval_state,provenance,
 approval_revision,source_evidence_sha256,revision,setup_completed_at,submitted_at,
 decision_at,decision_actor_id,created_at,updated_at)
VALUES('018f47f4-7b1b-7000-8000-000000000020',
 '51000000-0000-4000-8000-000000000002','52000000-0000-4000-8000-000000000002',
 'portal','completed','approved','native_v2',1,decode(repeat('45',32),'hex'),1,
 statement_timestamp(),statement_timestamp(),statement_timestamp(),
	'51000000-0000-4000-8000-000000000001',statement_timestamp(),statement_timestamp())`)
}

func assertIAMPortalProjection(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	userID, status, accountType, revision, err := queryIAMPortalProjection(
		ctx, database, 1, `{018f47f4-7b1b-7000-8000-000000000023}`)
	if err != nil || userID != iamUserTwoID || status != "ACTIVE" ||
		accountType != "portal" || revision != 1 {
		t.Fatalf("portal projection=%s/%s/%s/%d: %v",
			userID, status, accountType, revision, err)
	}
	mustExecIAMTenant(t, ctx, database, "t_iam_a", `UPDATE iam_users SET status='suspended',
revision=revision+1,updated_at=clock_timestamp() WHERE id=$1`, iamUserTwoID)
	userID, status, accountType, revision, err = queryIAMPortalProjection(
		ctx, database, 1, `{018f47f4-7b1b-7000-8000-000000000023}`)
	if err != nil || status != "INACTIVE" || revision != 2 {
		t.Fatalf("suspended portal projection=%s/%d: %v", status, revision, err)
	}
	_, _, _, _, err = queryIAMPortalProjection(
		ctx, database, 2, `{018f47f4-7b1b-7000-8000-000000000023}`)
	assertIAMSQLState(t, err, "42501")
	_, _, _, _, err = queryIAMPortalProjection(ctx, database, 1,
		`{018f47f4-7b1b-7000-8000-000000000023,018f47f4-7b1b-7000-8000-000000000023}`)
	assertIAMSQLState(t, err, "42501")
}

func queryIAMPortalProjection(
	ctx context.Context, database *sql.DB, clock int64, recordIDs string,
) (string, string, string, int64, error) {
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, "t_iam_a")
	if err != nil {
		return "", "", "", 0, err
	}
	defer connection.Close()
	query := `SELECT portal_user_id,status,account_type,result_revision
FROM iam_portal_subject_links_read_v1(
 't_iam_a'::name,$1::uuid,$2::uuid,1::bigint,1::bigint,
 $3::bigint,$4::uuid,$5::uuid,$6::uuid[])`
	var userID, status, accountType string
	var revision int64
	err = transaction.QueryRowContext(ctx, query, iamFoundationTenants[0].id,
		iamUserOneID, clock, "018f47f4-7b1b-7000-8000-000000000021",
		"018f47f4-7b1b-7000-8000-000000000022", recordIDs).Scan(
		&userID, &status, &accountType, &revision)
	if err != nil {
		_ = transaction.Rollback()
		return "", "", "", 0, err
	}
	if err = transaction.Commit(); err != nil {
		return "", "", "", 0, err
	}
	return userID, status, accountType, revision, nil
}
