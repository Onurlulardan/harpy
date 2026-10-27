//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

type billingRefreshTarget struct {
	found, ready                           bool
	operationID, licenseID, organizationID uuid.UUID
	environmentID                          uuid.UUID
	delivery                               string
	revision                               int64
	retryAt, databaseNow                   time.Time
}

func assertBillingLicenseRefreshTarget(
	t *testing.T, ctx context.Context, database *sql.DB,
	state billingLicenseAcceptedState,
) {
	t.Helper()
	workers := []*sql.Conn{
		billingLicenseProfileConnection(t, ctx, database, "core-worker"),
		billingLicenseProfileConnection(t, ctx, database, "core-worker"),
	}
	initial, err := acquireBillingRefreshTarget(ctx, workers[0])
	if err != nil || !initial.found || initial.ready ||
		initial.licenseID != state.fixture.licenseID {
		t.Fatalf("initial refresh target=%+v err=%v", initial, err)
	}
	forceBillingRefreshTargetDue(t, ctx, database, state.fixture)
	results := make(chan billingRefreshTarget, 2)
	errors := make(chan error, 2)
	for _, worker := range workers {
		go func(connection *sql.Conn) {
			target, acquireErr := acquireBillingRefreshTarget(ctx, connection)
			results <- target
			errors <- acquireErr
		}(worker)
	}
	left, right := <-results, <-results
	if err = <-errors; err != nil {
		t.Fatal(err)
	}
	if err = <-errors; err != nil {
		t.Fatal(err)
	}
	if left.operationID != right.operationID || left.ready == right.ready {
		t.Fatalf("raced refresh targets left=%+v right=%+v", left, right)
	}
	active := left
	if right.ready {
		active = right
	}
	claim, err := claimBillingLicense(ctx, workers[0], state.fixture, active.operationID, 30)
	if err != nil {
		t.Fatal(err)
	}
	assertLicenseAcceptOutcome(
		t, ctx, workers[0], state.fixture, claim, state.snapshot, "noop",
	)
	forceBillingRefreshTargetDue(t, ctx, database, state.fixture)
	rotated, err := acquireBillingRefreshTarget(ctx, workers[0])
	if err != nil || !rotated.ready || rotated.operationID == active.operationID {
		t.Fatalf("rotated refresh target=%+v err=%v", rotated, err)
	}
	replay, err := acquireBillingRefreshTarget(ctx, workers[1])
	if err != nil || replay.ready || replay.operationID != rotated.operationID {
		t.Fatalf("stable refresh target=%+v err=%v", replay, err)
	}
	assertBillingLicenseSaaSDenied(t, ctx, database, workers[0])
}

func billingLicenseProfileConnection(
	t *testing.T, ctx context.Context, database *sql.DB, profile string,
) *sql.Conn {
	t.Helper()
	connection := billingRuntimeConnection(t, ctx, database)
	if _, err := connection.ExecContext(
		ctx, "SET application_name = 'harpy-v2:"+profile+"'",
	); err != nil {
		t.Fatal(err)
	}
	return connection
}

func acquireBillingRefreshTarget(
	ctx context.Context, connection *sql.Conn,
) (billingRefreshTarget, error) {
	target := billingRefreshTarget{}
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.billing_license_refresh_target_acquire_v1()`).Scan(
		&target.found, &target.ready, &target.operationID,
		&target.licenseID, &target.organizationID, &target.environmentID,
		&target.delivery, &target.revision, &target.retryAt, &target.databaseNow,
	)
	return target, err
}

func forceBillingRefreshTargetDue(
	t *testing.T, ctx context.Context, database *sql.DB, fixture billingLicenseFixture,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE platform.billing_license_refresh_targets
SET next_attempt_at=clock_timestamp()-interval '1 second', lease_until=NULL,
    updated_at=clock_timestamp()-interval '2 seconds'
WHERE license_id=$1 AND environment_id=$2`, fixture.licenseID, fixture.environmentID)
	if err != nil {
		t.Fatal(err)
	}
}

func assertBillingLicenseSaaSDenied(
	t *testing.T, ctx context.Context, database *sql.DB, worker *sql.Conn,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `UPDATE platform.tenancy_deployment_policy
SET deployment_mode='saas', maximum_customer_tenants=NULL`); err != nil {
		t.Fatal(err)
	}
	if _, err := acquireBillingRefreshTarget(ctx, worker); billingPGCode(err) != "42501" {
		t.Fatalf("SaaS refresh target code=%q err=%v", billingPGCode(err), err)
	}
	operator := billingLicenseProfileConnection(
		t, ctx, database, "core-operator-license-reconcile",
	)
	var value int
	err := operator.QueryRowContext(ctx, `SELECT 1 FROM
platform.billing_license_restore_binding_v1($1)`, billingEnvironmentID).Scan(&value)
	if billingPGCode(err) != "42501" {
		t.Fatalf("SaaS restore binding code=%q err=%v", billingPGCode(err), err)
	}
	if _, err = database.ExecContext(ctx, `UPDATE platform.tenancy_deployment_policy
SET deployment_mode='onpremise', maximum_customer_tenants=0`); err != nil {
		t.Fatal(err)
	}
}

func assertBillingRestoreBindingAmbiguous(
	t *testing.T, ctx context.Context, database *sql.DB, connection *sql.Conn,
	state billingLicenseAcceptedState,
) {
	t.Helper()
	second := newBillingLicenseFixture()
	claim, err := claimBillingLicense(ctx, connection, second, billingV7(t), 30)
	if err != nil {
		t.Fatal(err)
	}
	assertLicenseAcceptOutcome(
		t, ctx, connection, second, claim,
		newBillingLicenseSnapshot(1, "active", "ambiguous"), "accepted",
	)
	operator := billingLicenseProfileConnection(
		t, ctx, database, "core-operator-license-reconcile",
	)
	var value int
	err = operator.QueryRowContext(ctx, `SELECT 1 FROM
platform.billing_license_restore_binding_v1($1)`, state.fixture.environmentID).Scan(&value)
	if billingPGCode(err) != "55000" {
		t.Fatalf("ambiguous restore binding code=%q err=%v", billingPGCode(err), err)
	}
}
