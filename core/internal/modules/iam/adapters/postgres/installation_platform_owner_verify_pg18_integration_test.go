//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const (
	ownerVerifierMigration = "20260729100048_create_iam_installation_instance_owner_verify_v1"
	ownerVerifierOperation = "56000000-0000-4000-8000-000000000001"
	ownerVerifierSubject   = "51000000-0000-4000-8000-000000000001"
	ownerVerifierRole      = "52000000-0000-4000-8000-000000000001"
	ownerVerifierGrant     = "54000000-0000-4000-8000-000000000001"
	ownerVerifierBinding   = "53000000-0000-4000-8000-000000000001"
	ownerVerifierRequest   = "61616161616161616161616161616161" +
		"61616161616161616161616161616161"
	ownerVerifierWrongRequest = "62626262626262626262626262626262" +
		"62626262626262626262626262626262"
)

const ownerVerifierCall = `SELECT platform.iam_installation_instance_owner_verify_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid,
    $6::uuid,$7::uuid,$8::uuid,$9::bigint,$10::text)`

func TestIAMInstallationPlatformOwnerVerifyPG18(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			database := startIAMPlatformGuardPG18(t, ctx, "owner_verify_"+origin)
			fixture := prepareOwnerTransferPG18(t, ctx, database, origin)
			applyIAMPlatformGuardDocument(t, ctx, fixture.migrator,
				ownerVerifierMigration,
				readIAMPlatformAuthMigration(t, origin, ownerVerifierMigration))
			seedInstallationOwnerVerifierReceiptPG18(t, ctx, fixture)
			setInstallationOwnerVerifierProfilePG18(t, ctx, fixture.runtime,
				"harpy-v2:core-operator-initialize")
			assertInstallationOwnerVerifierACLPG18(t, ctx, fixture)

			before := installationOwnerVerifierSnapshotPG18(t, ctx, fixture)
			verifyInstallationOwnerReadOnlyPG18(t, ctx, fixture.runtime)
			assertInstallationOwnerVerifierNegativePG18(t, ctx, fixture)
			after := installationOwnerVerifierSnapshotPG18(t, ctx, fixture)
			if before != after {
				t.Fatal("installation owner verification mutated authority state")
			}

			exerciseOwnerTransferCommit(t, ctx, fixture, firstOwnerTransferAttempt())
			transferred := installationOwnerVerifierSnapshotPG18(t, ctx, fixture)
			verifyInstallationOwnerReadOnlyPG18(t, ctx, fixture.runtime)
			if installationOwnerVerifierSnapshotPG18(t, ctx, fixture) != transferred {
				t.Fatal("post-transfer owner verification mutated authority state")
			}
		})
	}
}

func seedInstallationOwnerVerifierReceiptPG18(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
) {
	t.Helper()
	_, err := fixture.migrator.ExecContext(ctx, `
INSERT INTO platform.iam_platform_mutation_receipts (
    operation_id,environment_id,mutation_kind,request_sha256,
    actor_user_id,target_user_id,subject_id,role_id,subject_role_binding_id,
    role_permission_binding_id,permission_id,target_authorization_version,
    platform_clock,occurred_at
) SELECT $1,$2,'initial_owner',decode($3,'hex'),$4,$4,$5,$6,$7,$8,
         permission.id,1,1,clock_timestamp()
FROM platform.iam_platform_permissions AS permission
WHERE permission.permission_key='environment_all'`,
		ownerVerifierOperation, ownerEnvID, ownerVerifierRequest, ownerUserID,
		ownerVerifierSubject, ownerVerifierRole, ownerVerifierBinding, ownerVerifierGrant)
	if err != nil {
		t.Fatalf("seed installation owner receipt: %v", err)
	}
}

type installationOwnerVerifierQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func ownerVerifierArguments(request string) []any {
	return []any{
		ownerVerifierOperation, ownerEnvID, ownerTenantID, ownerUserID,
		ownerVerifierSubject, ownerVerifierRole, ownerVerifierGrant,
		ownerVerifierBinding, int64(1), request,
	}
}

func callInstallationOwnerVerifierPG18(
	ctx context.Context, querier installationOwnerVerifierQuerier, request string,
) (bool, error) {
	var verified bool
	err := querier.QueryRowContext(ctx, ownerVerifierCall,
		ownerVerifierArguments(request)...).Scan(&verified)
	return verified, err
}

func verifyInstallationOwnerReadOnlyPG18(
	t *testing.T, ctx context.Context, runtime *sql.Conn,
) {
	t.Helper()
	transaction, err := runtime.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	verified, err := callInstallationOwnerVerifierPG18(ctx, transaction, ownerVerifierRequest)
	if err != nil || !verified {
		_ = transaction.Rollback()
		t.Fatalf("read-only installation owner verification = %t/%v", verified, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit read-only installation owner verification: %v", err)
	}
}

func assertInstallationOwnerVerifierNegativePG18(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
) {
	t.Helper()
	_, err := callInstallationOwnerVerifierPG18(
		ctx, fixture.runtime, ownerVerifierWrongRequest)
	assertIAMSQLState(t, err, "55000")
	setInstallationOwnerVerifierProfilePG18(t, ctx, fixture.runtime,
		"harpy-v2:core-api")
	_, err = callInstallationOwnerVerifierPG18(
		ctx, fixture.runtime, ownerVerifierRequest)
	assertIAMSQLState(t, err, "42501")
	setInstallationOwnerVerifierProfilePG18(t, ctx, fixture.runtime,
		"harpy-v2:core-operator-initialize")
	setInstallationOwnerVerifierProfilePG18(t, ctx, fixture.migrator,
		"harpy-v2:core-operator-initialize")
	_, err = callInstallationOwnerVerifierPG18(
		ctx, fixture.migrator, ownerVerifierRequest)
	assertIAMSQLState(t, err, "42501")
}

func setInstallationOwnerVerifierProfilePG18(
	t *testing.T, ctx context.Context, connection *sql.Conn, application string,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx,
		"SELECT pg_catalog.set_config('application_name',$1,false)", application); err != nil {
		t.Fatal(err)
	}
}
