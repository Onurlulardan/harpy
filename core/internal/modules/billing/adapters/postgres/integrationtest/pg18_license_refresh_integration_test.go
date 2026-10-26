//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

type billingLicenseAcceptedState struct {
	fixture  billingLicenseFixture
	snapshot billingLicenseSnapshot
	revision int64
}

func assertBillingLicenseLifecycle(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	assertBillingLicenseClock(t, ctx, database)
	connection := billingRuntimeConnection(t, ctx, database)
	state := assertBillingLicenseRefresh(t, ctx, database, connection)
	assertBillingLicenseRefreshTarget(t, ctx, database, state)
	state = assertBillingLicenseRestore(t, ctx, database, connection, state)
	assertBillingLicenseOffering(t, ctx, connection, state)
	assertBillingLicenseRuntime(t, ctx, connection, state)
	assertBillingLicenseEvidenceImmutable(t, ctx, connection)
	assertBillingPlatformLicenseAPI(t, ctx, database, state)
	assertBillingRestoreBindingAmbiguous(t, ctx, database, connection, state)
}

func assertBillingLicenseRefresh(
	t *testing.T, ctx context.Context, database *sql.DB, connection *sql.Conn,
) billingLicenseAcceptedState {
	t.Helper()
	fixture := newBillingLicenseFixture()
	operationID := billingV7(t)
	first, err := claimBillingLicense(ctx, connection, fixture, operationID, 5)
	if err != nil || len(first.nonce) != 32 {
		t.Fatalf("first license claim: %+v %v", first, err)
	}
	if _, err = claimBillingLicense(ctx, connection, fixture, operationID, 5); billingPGCode(err) != "55P03" {
		t.Fatalf("active license lease code=%q err=%v", billingPGCode(err), err)
	}
	if _, err = connection.ExecContext(ctx, "SELECT pg_catalog.pg_sleep(5.2)"); err != nil {
		t.Fatal(err)
	}
	reclaimed, err := claimBillingLicense(ctx, connection, fixture, operationID, 30)
	if err != nil || reclaimed.leaseToken == first.leaseToken ||
		!bytes.Equal(reclaimed.nonce, first.nonce) ||
		reclaimed.databaseNow.IsZero() ||
		reclaimed.leaseUntil.Sub(reclaimed.databaseNow) != 30*time.Second {
		t.Fatalf("reclaimed license claim: %+v %v", reclaimed, err)
	}
	var expired string
	if err = connection.QueryRowContext(ctx, `SELECT state FROM
platform.billing_license_refresh_attempts WHERE operation_id=$1 AND attempt=1`,
		operationID).Scan(&expired); err != nil || expired != "lease_expired" {
		t.Fatalf("expired attempt=%q err=%v", expired, err)
	}
	initial := newBillingLicenseSnapshot(2, "active", "initial")
	assertLicenseAcceptOutcome(t, ctx, connection, fixture, reclaimed, initial, "accepted")
	assertLicenseNoopAndFences(t, ctx, connection, fixture, initial)
	assertLicenseRefreshFailures(t, ctx, connection, fixture)
	terminal := newBillingLicenseSnapshot(3, "revoked", "terminal")
	claim, err := claimBillingLicense(ctx, connection, fixture, billingV7(t), 30)
	if err != nil {
		t.Fatal(err)
	}
	assertLicenseAcceptOutcome(t, ctx, connection, fixture, claim, terminal, "accepted")
	winner := assertConcurrentLicenseCAS(t, ctx, database, connection, fixture)
	var revision int64
	if err = connection.QueryRowContext(ctx, `SELECT revision FROM
platform.billing_license_heads WHERE license_id=$1 AND environment_id=$2`,
		fixture.licenseID, fixture.environmentID).Scan(&revision); err != nil || revision != 3 {
		t.Fatalf("license head revision=%d err=%v", revision, err)
	}
	return billingLicenseAcceptedState{fixture: fixture, snapshot: winner, revision: revision}
}

func assertLicenseNoopAndFences(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	fixture billingLicenseFixture, initial billingLicenseSnapshot,
) {
	t.Helper()
	before := countBillingLicenseRows(t, ctx, connection, "billing_license_snapshots")
	claim, _ := claimBillingLicense(ctx, connection, fixture, billingV7(t), 30)
	assertLicenseAcceptOutcome(t, ctx, connection, fixture, claim, initial, "noop")
	if got := countBillingLicenseRows(t, ctx, connection, "billing_license_snapshots"); got != before {
		t.Fatalf("noop appended snapshot: before=%d after=%d", before, got)
	}
	conflict := newBillingLicenseSnapshot(2, "active", "conflict")
	claim, _ = claimBillingLicense(ctx, connection, fixture, billingV7(t), 30)
	assertLicenseAcceptOutcome(t, ctx, connection, fixture, claim, conflict, "sequence_conflict")
	rollback := newBillingLicenseSnapshot(1, "active", "rollback")
	claim, _ = claimBillingLicense(ctx, connection, fixture, billingV7(t), 30)
	assertLicenseAcceptOutcome(t, ctx, connection, fixture, claim, rollback, "sequence_rollback")
}

func assertLicenseRefreshFailures(
	t *testing.T, ctx context.Context, connection *sql.Conn, fixture billingLicenseFixture,
) {
	t.Helper()
	headBefore := countBillingLicenseRows(t, ctx, connection, "billing_license_heads")
	snapshotsBefore := countBillingLicenseRows(t, ctx, connection, "billing_license_snapshots")
	for _, kind := range []string{"source_unavailable", "verification_failed"} {
		claim, err := claimBillingLicense(ctx, connection, fixture, billingV7(t), 30)
		if err != nil {
			t.Fatal(err)
		}
		state, err := failBillingLicense(ctx, connection, claim, kind)
		if err != nil || state != "pending" {
			t.Fatalf("license %s state=%q err=%v", kind, state, err)
		}
		var attemptState, failureKind string
		err = connection.QueryRowContext(ctx, `SELECT state,failure_kind FROM
platform.billing_license_refresh_attempts WHERE operation_id=$1 AND attempt=1`,
			claim.operationID).Scan(&attemptState, &failureKind)
		wantState := map[string]string{
			"source_unavailable": "source_failed", "verification_failed": "verification_failed",
		}[kind]
		if err != nil || attemptState != wantState || failureKind != kind {
			t.Fatalf("license failure evidence=%s/%s err=%v", attemptState, failureKind, err)
		}
	}
	if countBillingLicenseRows(t, ctx, connection, "billing_license_heads") != headBefore ||
		countBillingLicenseRows(t, ctx, connection, "billing_license_snapshots") != snapshotsBefore {
		t.Fatal("retryable license failure replaced the accepted cache")
	}
}

func assertConcurrentLicenseCAS(
	t *testing.T, ctx context.Context, database *sql.DB, connection *sql.Conn,
	fixture billingLicenseFixture,
) billingLicenseSnapshot {
	t.Helper()
	left, _ := claimBillingLicense(ctx, connection, fixture, billingV7(t), 30)
	right, _ := claimBillingLicense(ctx, connection, fixture, billingV7(t), 30)
	snapshots := []billingLicenseSnapshot{
		newBillingLicenseSnapshot(4, "active", "cas-left"),
		newBillingLicenseSnapshot(4, "active", "cas-right"),
	}
	claims := []billingLicenseClaim{left, right}
	workers := []*sql.Conn{
		billingRuntimeConnection(t, ctx, database),
		billingRuntimeConnection(t, ctx, database),
	}
	type result struct {
		outcome string
		err     error
	}
	results := make(chan result, 2)
	for index := range claims {
		go func(index int) {
			outcome, err := acceptBillingLicense(
				ctx, workers[index], fixture, claims[index], snapshots[index],
			)
			results <- result{outcome: outcome, err: err}
		}(index)
	}
	seen := map[string]int{}
	for range claims {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		seen[result.outcome]++
	}
	if seen["accepted"] != 1 || seen["sequence_conflict"] != 1 {
		t.Fatalf("concurrent license outcomes=%v", seen)
	}
	var digest string
	if err := connection.QueryRowContext(ctx, `SELECT payload_sha256 FROM
platform.billing_license_heads WHERE license_id=$1 AND environment_id=$2`,
		fixture.licenseID, fixture.environmentID).Scan(&digest); err != nil {
		t.Fatal(err)
	}
	if snapshots[0].digest == digest {
		return snapshots[0]
	}
	return snapshots[1]
}

func assertLicenseAcceptOutcome(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	fixture billingLicenseFixture, claim billingLicenseClaim,
	snapshot billingLicenseSnapshot, expected string,
) {
	t.Helper()
	outcome, err := acceptBillingLicense(ctx, connection, fixture, claim, snapshot)
	if err != nil || outcome != expected {
		t.Fatalf("accept license outcome=%q want=%q err=%v", outcome, expected, err)
	}
}
