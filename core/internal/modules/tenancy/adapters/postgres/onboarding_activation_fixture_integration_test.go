//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func activateOnboarding(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	fixture onboardingFixture,
	expectedRevision int64,
) (int64, string) {
	t.Helper()
	var operationID, tenantID, migrationReceiptID uuid.UUID
	var kind, stepSHA, migrationSHA, ackSHA sql.NullString
	var revision int64
	var databaseNow, leaseUntil time.Time
	err := transaction.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_onboarding_activate_v1($1,$2,$3)`, fixture.operationID,
		strings.Repeat("a", 64), expectedRevision).Scan(&operationID, &tenantID, &kind,
		&stepSHA, &migrationReceiptID, &migrationSHA, &ackSHA, &revision,
		&databaseNow, &leaseUntil)
	if err != nil {
		t.Fatal(err)
	}
	if operationID != fixture.operationID || tenantID != fixture.tenantID ||
		kind.String != fixture.kind || stepSHA.String != fixture.stepSHA ||
		migrationReceiptID == uuid.Nil || !migrationSHA.Valid || !leaseUntil.After(databaseNow) {
		t.Fatalf("activation claim identity=%s tenant=%s revision=%d", operationID, tenantID, revision)
	}
	return revision, ackSHA.String
}

func completeOnboarding(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	fixture onboardingFixture,
	receiptID uuid.UUID,
	expectedRevision int64,
) {
	t.Helper()
	_, err := transaction.ExecContext(ctx, `SELECT
platform.tenancy_onboarding_complete_v1($1,$2,$3,$4,$5)`, fixture.operationID,
		receiptID, strings.Repeat("d", 64), strings.Repeat("e", 64), expectedRevision)
	if err != nil {
		t.Fatal(err)
	}
}

func onboardingState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	operationID uuid.UUID,
) (string, int64) {
	t.Helper()
	var state string
	var revision int64
	if err := database.QueryRowContext(ctx, `SELECT state, revision FROM
platform.tenancy_onboarding_operations WHERE operation_id=$1`, operationID).
		Scan(&state, &revision); err != nil {
		t.Fatal(err)
	}
	return state, revision
}
