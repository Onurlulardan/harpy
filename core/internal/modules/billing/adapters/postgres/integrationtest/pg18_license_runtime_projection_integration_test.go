//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	licensepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/licenseauthority"
	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

type billingRuntimeGrant struct {
	state, key, valueType string
	boolean               sql.NullBool
	integer               sql.NullInt64
	unlimited             bool
}

func assertBillingLicenseRuntime(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	state billingLicenseAcceptedState,
) {
	t.Helper()
	assertBillingRuntimeState(t, ctx, connection, state.fixture, "valid")
	grace := newBillingLicenseSnapshot(6, "active", "grace")
	grace.grantUntil = time.Now().UTC().Add(-time.Minute)
	grace.graceUntil = time.Now().UTC().Add(time.Hour)
	claim, err := claimBillingLicense(ctx, connection, state.fixture, billingV7(t), 30)
	if err != nil {
		t.Fatal(err)
	}
	assertLicenseAcceptOutcome(t, ctx, connection, state.fixture, claim, grace, "accepted")
	assertBillingRuntimeState(t, ctx, connection, state.fixture, "grace")
	expired := newBillingLicenseSnapshot(7, "active", "expired")
	now := time.Now().UTC()
	expired.issuedAt, expired.notBefore = now.Add(-3*time.Hour), now.Add(-3*time.Hour)
	expired.refreshAfter, expired.grantUntil = now.Add(-150*time.Minute), now.Add(-2*time.Hour)
	expired.graceUntil = now.Add(-time.Hour)
	claim, err = claimBillingLicense(ctx, connection, state.fixture, billingV7(t), 30)
	if err != nil {
		t.Fatal(err)
	}
	assertLicenseAcceptOutcome(t, ctx, connection, state.fixture, claim, expired, "accepted")
	assertBillingRuntimeState(t, ctx, connection, state.fixture, "build_off")
	terminal := newBillingLicenseSnapshot(8, "suspended", "terminal-now")
	claim, err = claimBillingLicense(ctx, connection, state.fixture, billingV7(t), 30)
	if err != nil {
		t.Fatal(err)
	}
	assertLicenseAcceptOutcome(t, ctx, connection, state.fixture, claim, terminal, "accepted")
	assertBillingRuntimeState(t, ctx, connection, state.fixture, "build_off")
	var status string
	if err = connection.QueryRowContext(ctx, `SELECT snapshot.status FROM
platform.billing_license_heads AS head JOIN platform.billing_license_snapshots AS snapshot
ON snapshot.snapshot_id=head.snapshot_id WHERE head.license_id=$1 AND head.environment_id=$2`,
		state.fixture.licenseID, state.fixture.environmentID).Scan(&status); err != nil || status != "suspended" {
		t.Fatalf("terminal license status=%q err=%v", status, err)
	}
}

func assertBillingRuntimeState(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	fixture billingLicenseFixture, expected string,
) {
	t.Helper()
	rows, err := connection.QueryContext(ctx, `SELECT runtime_state, entitlement_key,
value_type, boolean_value, integer_value, unlimited
FROM platform.billing_license_runtime_entitlements_v1($1)`, fixture.environmentID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		grant := billingRuntimeGrant{}
		if err = rows.Scan(&grant.state, &grant.key, &grant.valueType,
			&grant.boolean, &grant.integer, &grant.unlimited); err != nil {
			t.Fatal(err)
		}
		count++
		if grant.state != expected {
			t.Fatalf("runtime state=%q want=%q", grant.state, expected)
		}
		assertBillingRuntimeGrant(t, grant, expected)
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 13 {
		t.Fatalf("runtime grant count=%d", count)
	}
	assertBillingTypedRuntimeRead(t, ctx, connection, fixture, expected)
}

func assertBillingRuntimeGrant(t *testing.T, grant billingRuntimeGrant, state string) {
	t.Helper()
	if state != "build_off" {
		if grant.valueType == "boolean" && (!grant.boolean.Valid || !grant.boolean.Bool) {
			t.Fatalf("runtime boolean grant=%+v", grant)
		}
		if grant.key == "iam.active_users" && (grant.integer.Valid || !grant.unlimited) {
			t.Fatalf("runtime unlimited grant=%+v", grant)
		}
		if grant.valueType == "integer" && grant.key != "iam.active_users" &&
			(!grant.integer.Valid || grant.integer.Int64 != 10 || grant.unlimited) {
			t.Fatalf("runtime integer grant=%+v", grant)
		}
		return
	}
	if grant.valueType == "integer" {
		if !grant.integer.Valid || grant.integer.Int64 != 0 || grant.unlimited {
			t.Fatalf("build-off integer grant=%+v", grant)
		}
		return
	}
	want := grant.key == "workflow.enabled" || grant.key == "api.access"
	if !grant.boolean.Valid || grant.boolean.Bool != want {
		t.Fatalf("build-off boolean grant=%+v want=%t", grant, want)
	}
}

func assertBillingTypedRuntimeRead(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	fixture billingLicenseFixture, expected string,
) {
	t.Helper()
	storage, err := licensepostgres.NewStore(licensepostgres.DatabaseAccess{
		QueryRow: func(ctx context.Context, statement string, arguments ...any) licensepostgres.Row {
			return connection.QueryRowContext(ctx, statement, arguments...)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	reader, err := licenseapp.NewRuntimeReader(storage)
	if err != nil {
		t.Fatal(err)
	}
	result, err := reader.ReadLicenseRuntimeEntitlements(
		ctx, fixture.environmentID.String(),
	)
	if err != nil || string(result.State) != expected ||
		len(result.Projection.Boolean) != 6 || len(result.Projection.Limits) != 7 ||
		result.Projection.Channel != fixture.channel || result.DatabaseNow.IsZero() {
		t.Fatalf("typed runtime projection=%+v err=%v", result, err)
	}
}
