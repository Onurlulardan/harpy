//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

func TestInstallationCredentialDeliveryLifecycleAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := onboardingTestDatabase(t, ctx, "onpremise")
	fixture := newOnboardingFixture(t, "onpremise")
	seedSucceededOnboarding(t, ctx, database, fixture)

	firstAttempt, firstCredential := uuid.New(), uuid.New()
	revision := stageInstallationDelivery(t, ctx, database, fixture,
		firstAttempt, firstCredential)
	claimToken := strings.Repeat("a", 64)
	revision = claimInstallationDelivery(t, ctx, database, fixture.operationID,
		firstAttempt, revision, "local-terminal", claimToken)
	revision = unknownInstallationDelivery(t, ctx, database, fixture.operationID,
		firstAttempt, revision, "local-terminal", claimToken)

	secondAttempt, secondCredential := uuid.New(), uuid.New()
	recoveryToken := strings.Repeat("b", 64)
	revision = recoverInstallationDelivery(t, ctx, database, fixture.operationID,
		firstAttempt, firstCredential, secondAttempt, secondCredential, revision,
		recoveryToken)
	revision = recordInstallationRecovery(t, ctx, database, fixture.operationID,
		firstAttempt, firstCredential, secondAttempt, secondCredential, revision,
		recoveryToken)

	secondClaim := strings.Repeat("c", 64)
	revision = claimInstallationDelivery(t, ctx, database, fixture.operationID,
		secondAttempt, revision, "local-terminal", secondClaim)
	revision = acknowledgeInstallationDelivery(t, ctx, database, fixture.operationID,
		secondAttempt, revision, "local-terminal", secondClaim)
	revision = transitionInstallationDelivery(t, ctx, database,
		"tenancy_installation_delivery_secret_deleted_v1",
		fixture.operationID, secondAttempt, revision, "secret_deleted")
	_ = transitionInstallationDelivery(t, ctx, database,
		"tenancy_installation_delivery_complete_v1",
		fixture.operationID, secondAttempt, revision, "completed")
	assertInstallationDeliveryHistory(t, ctx, database, fixture.operationID)
	assertInstallationDeliveryAuthority(t, ctx, database, fixture, secondAttempt,
		secondCredential)
}

func seedSucceededOnboarding(
	t *testing.T, ctx context.Context, database *sql.DB, fixture onboardingFixture,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	acceptOnboarding(t, ctx, transaction, fixture, fixture.manifestSHA)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `UPDATE platform.tenancy_onboarding_operations
SET state='succeeded',revision=2,updated_at=statement_timestamp()
WHERE operation_id=$1`, fixture.operationID); err != nil {
		t.Fatal(err)
	}
}

func stageInstallationDelivery(
	t *testing.T, ctx context.Context, database *sql.DB, fixture onboardingFixture,
	attemptID, credentialID uuid.UUID,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var state string
	var revision int64
	err := transaction.QueryRowContext(ctx, `SELECT delivery_state,delivery_revision
FROM platform.tenancy_installation_delivery_stage_v1($1,$2,$3,1,$4,$5)`,
		fixture.operationID, attemptID, credentialID, fixture.manifestSHA,
		credentialdeliveryv1.ManifestSHA256()).Scan(&state, &revision)
	if err != nil || state != "ready_to_display" || revision != 1 {
		t.Fatalf("stage state=%s revision=%d err=%v", state, revision, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return revision
}

func claimInstallationDelivery(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID, attemptID uuid.UUID, revision int64, adapter, token string,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var state string
	var next int64
	err := transaction.QueryRowContext(ctx, `SELECT delivery_state,delivery_revision
FROM platform.tenancy_installation_delivery_claim_v1($1,$2,$3,$4,$5,60)`,
		operationID, attemptID, revision, adapter, token).Scan(&state, &next)
	if err != nil || state != "claimed" || next != revision+1 {
		t.Fatalf("claim state=%s revision=%d err=%v", state, next, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return next
}

func unknownInstallationDelivery(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID, attemptID uuid.UUID, revision int64, adapter, token string,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var state string
	var next int64
	err := transaction.QueryRowContext(ctx, `SELECT delivery_state,delivery_revision
FROM platform.tenancy_installation_delivery_unknown_v1($1,$2,$3,$4,$5,'outcome_unknown')`,
		operationID, attemptID, revision, adapter, token).Scan(&state, &next)
	if err != nil || state != "outcome_unknown" || next != revision+1 {
		t.Fatalf("unknown state=%s revision=%d err=%v", state, next, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return next
}
