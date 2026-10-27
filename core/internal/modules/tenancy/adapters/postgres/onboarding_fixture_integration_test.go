//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/domain"
	"github.com/google/uuid"
)

type onboardingFixture struct {
	operationID, environmentID, tenantID, firstUserID uuid.UUID
	organizationID, adminBindingID, ownerBindingID    uuid.UUID
	offeringID, registryBindingID                     uuid.UUID
	schema, kind, mode, purpose                       string
	requestSHA, manifestSHA, stepSHA                  string
	stateSHA, proposalSHA                             string
	createdAt                                         time.Time
}

const onboardingAcceptSQL = `SELECT * FROM platform.tenancy_onboarding_accept_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31)`

func newOnboardingFixture(t *testing.T, mode string) onboardingFixture {
	t.Helper()
	tenantID := uuid.New()
	schema, err := domain.SchemaNameFromUUID(tenantID.String())
	if err != nil {
		t.Fatal(err)
	}
	result := onboardingFixture{operationID: uuid.New(),
		environmentID: uuid.MustParse(testEnvironmentID), tenantID: tenantID,
		firstUserID: uuid.New(), organizationID: uuid.New(), adminBindingID: uuid.New(),
		ownerBindingID: uuid.New(), offeringID: uuid.New(), registryBindingID: uuid.New(),
		schema: schema.String(), kind: "environment_initialize", mode: mode,
		requestSHA: strings.Repeat("1", 64), manifestSHA: strings.Repeat("2", 64),
		stepSHA: strings.Repeat("3", 64), createdAt: time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)}
	if mode == "onpremise" {
		result.purpose = "onpremise_primary"
	} else {
		result.purpose = "saas_control"
		result.stateSHA, result.proposalSHA = strings.Repeat("8", 64), strings.Repeat("9", 64)
	}
	return result
}

func onboardingTestDatabase(
	t *testing.T,
	ctx context.Context,
	mode string,
) *sql.DB {
	t.Helper()
	database := startTenantMigrationDatabase(t, ctx)
	resetTenantMigrationCatalog(t, ctx, database)
	runPlatformMigrationsThrough(t, ctx, database, "fresh", 20260729100036)
	insertDeploymentPolicy(t, ctx, database, mode)
	return database
}

func (fixture onboardingFixture) acceptArguments(manifest string) []any {
	var sourceSequence, sourceSHA, recoveryGeneration, recoverySHA any
	var stateSHA, proposalSHA any
	if fixture.mode == "onpremise" {
		sourceSequence, sourceSHA = int64(1), strings.Repeat("5", 64)
		recoveryGeneration, recoverySHA = int64(1), strings.Repeat("6", 64)
	} else {
		stateSHA, proposalSHA = fixture.stateSHA, fixture.proposalSHA
	}
	return []any{fixture.operationID, fixture.environmentID, fixture.kind, "initialize",
		fixture.requestSHA, manifest, fixture.stepSHA, fixture.mode, fixture.tenantID,
		fixture.schema, fixture.purpose, fixture.firstUserID, fixture.organizationID,
		fixture.adminBindingID, fixture.ownerBindingID, fixture.offeringID, int64(1),
		strings.Repeat("4", 64), fixture.createdAt, sourceSequence, sourceSHA,
		recoveryGeneration, recoverySHA, fixture.registryBindingID, "environment",
		fixture.environmentID, map[bool]int64{true: 0, false: 1}[fixture.mode == "onpremise"],
		nil, stateSHA, proposalSHA, fixture.createdAt}
}

func acceptOnboarding(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	fixture onboardingFixture,
	manifest string,
) (string, int64) {
	t.Helper()
	var operationID uuid.UUID
	var state string
	var revision int64
	var updatedAt time.Time
	arguments := fixture.acceptArguments(manifest)
	err := transaction.QueryRowContext(ctx, onboardingAcceptSQL, arguments...).
		Scan(&operationID, &state, &revision, &updatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if operationID != fixture.operationID || updatedAt.IsZero() {
		t.Fatalf("accept identity=%s updated=%s", operationID, updatedAt)
	}
	return state, revision
}

func claimOnboarding(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	fixture onboardingFixture,
	attemptID uuid.UUID,
	expectedRevision int64,
	leaseSeconds int,
) (int64, time.Time) {
	t.Helper()
	var operationID, returnedAttempt uuid.UUID
	var attemptNumber, revision int64
	var leaseUntil time.Time
	err := transaction.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_onboarding_claim_v2($1,$2,$3,$4,$5,$6)`, fixture.operationID,
		attemptID, "operator", strings.Repeat("a", 64), expectedRevision, leaseSeconds).
		Scan(&operationID, &returnedAttempt, &attemptNumber, &revision, &leaseUntil)
	if err != nil {
		t.Fatal(err)
	}
	if operationID != fixture.operationID || returnedAttempt != attemptID {
		t.Fatalf("claim identity=%s/%s", operationID, returnedAttempt)
	}
	return revision, leaseUntil
}

func recordMigration(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	fixture onboardingFixture,
	attemptID uuid.UUID,
	expectedRevision int64,
) (uuid.UUID, int64) {
	t.Helper()
	receiptID := uuid.New()
	var operationID uuid.UUID
	var state string
	var revision int64
	var updatedAt time.Time
	err := transaction.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_onboarding_record_v1($1,$2,$3,$4,$5,$6,$7,$8)`, receiptID,
		fixture.operationID, attemptID, "migration_completed", strings.Repeat("b", 64),
		strings.Repeat("c", 64), fixture.createdAt, expectedRevision).
		Scan(&operationID, &state, &revision, &updatedAt)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"onpremise": "migration_ready", "saas": "authority_pending"}[fixture.mode]
	if operationID != fixture.operationID || state != want {
		t.Fatalf("migration transition=%s/%s", operationID, state)
	}
	return receiptID, revision
}
