//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

func recoverInstallationDelivery(
	t *testing.T, ctx context.Context, database *sql.DB, operationID,
	currentAttempt, currentCredential, nextAttempt, nextCredential uuid.UUID,
	revision int64, token string,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var next int64
	var cause string
	err := transaction.QueryRowContext(ctx, `SELECT delivery_revision,recovery_cause
FROM platform.tenancy_installation_delivery_recover_v1(
$1,$2,$3,1,$4,$5,2,$6,$7,$8,60)`, operationID, currentAttempt,
		currentCredential, nextAttempt, nextCredential, revision, token,
		credentialdeliveryv1.ManifestSHA256()).Scan(&next, &cause)
	if err != nil || next != revision+1 || cause != "outcome_unknown" {
		t.Fatalf("recover revision=%d cause=%s err=%v", next, cause, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return next
}

func recordInstallationRecovery(
	t *testing.T, ctx context.Context, database *sql.DB, operationID,
	currentAttempt, currentCredential, nextAttempt, nextCredential uuid.UUID,
	revision int64, token string,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var state string
	var next int64
	err := transaction.QueryRowContext(ctx, `SELECT delivery_state,delivery_revision
FROM platform.tenancy_installation_delivery_record_v1(
$1,$2,$3,1,$4,$5,2,$6,$7,$8,'outcome_unknown')`, operationID,
		currentAttempt, currentCredential, nextAttempt, nextCredential, revision,
		token, credentialdeliveryv1.ManifestSHA256()).Scan(&state, &next)
	if err != nil || state != "ready_to_display" || next != revision+1 {
		t.Fatalf("record state=%s revision=%d err=%v", state, next, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return next
}

func acknowledgeInstallationDelivery(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID, attemptID uuid.UUID, revision int64, adapter, token string,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var state string
	var next int64
	err := transaction.QueryRowContext(ctx, `SELECT delivery_state,delivery_revision
FROM platform.tenancy_installation_delivery_ack_v1($1,$2,$3,$4,$5)`,
		operationID, attemptID, revision, adapter, token).Scan(&state, &next)
	if err != nil || state != "acknowledged" || next != revision+1 {
		t.Fatalf("ack state=%s revision=%d err=%v", state, next, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return next
}

func transitionInstallationDelivery(
	t *testing.T, ctx context.Context, database *sql.DB, routine string,
	operationID, attemptID uuid.UUID, revision int64, want string,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var state string
	var next int64
	query := `SELECT delivery_state,delivery_revision FROM platform.` + routine + `($1,$2,$3)`
	if err := transaction.QueryRowContext(ctx, query, operationID, attemptID, revision).
		Scan(&state, &next); err != nil || state != want || next != revision+1 {
		t.Fatalf("%s state=%s revision=%d", routine, state, next)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return next
}

func assertInstallationDeliveryHistory(
	t *testing.T, ctx context.Context, database *sql.DB, operationID uuid.UUID,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	rows, err := transaction.QueryContext(ctx, `SELECT credential_version,attempt_state,cause
FROM platform.tenancy_installation_delivery_attempts_v1($1)`, operationID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	states := map[int64]string{}
	for rows.Next() {
		var version int64
		var state string
		var cause sql.NullString
		if err = rows.Scan(&version, &state, &cause); err != nil {
			t.Fatal(err)
		}
		states[version] = state + ":" + cause.String
	}
	if states[1] != "superseded:outcome_unknown" || states[2] != "completed:" {
		t.Fatalf("attempt history=%v", states)
	}
}

func assertInstallationDeliveryAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, fixture onboardingFixture,
	attemptID, credentialID uuid.UUID,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `SELECT * FROM
platform.tenancy_installation_delivery_stage_v1($1,$2,$3,1,$4,$5)`,
		fixture.operationID, attemptID, credentialID, fixture.manifestSHA,
		strings.Repeat("d", 64))
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("owner invocation err=%v", err)
	}
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	_, err = transaction.ExecContext(ctx, `UPDATE
platform.tenancy_installation_credential_deliveries SET revision=99 WHERE operation_id=$1`,
		fixture.operationID)
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("runtime direct DML err=%v", err)
	}
}
