//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type billingUsageState struct {
	openings, periods, totals int
}

func assertBillingInstallationUsageCapability(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	fixture := createBillingUsageCapabilityFixture(t, ctx, database)
	for _, signature := range []string{
		"platform.billing_installation_open_usage_v2(uuid,uuid,text)",
	} {
		var allowed bool
		if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_function_privilege(
 'harpy_runtime',$1,'EXECUTE'
)`, signature).Scan(&allowed); err != nil || !allowed {
			t.Fatalf("Billing usage EXECUTE=%t signature=%s err=%v",
				allowed, signature, err)
		}
	}
	initial := readBillingUsageState(
		t, ctx, database, fixture.operationID, fixture.tenantID,
	)
	randomOperation := billingV7(t)
	for _, attempt := range []struct {
		name, manifest, code string
		tenant               uuid.UUID
		mint                 *uuid.UUID
	}{
		{"no_capability", fixture.manifestSHA256, "42501",
			fixture.tenantID, nil},
		{"wrong_operation", fixture.manifestSHA256, "42501",
			fixture.tenantID, &randomOperation},
		{"wrong_tenant", fixture.manifestSHA256, "55000",
			billingV7(t), &fixture.operationID},
		{"wrong_manifest", billingAcceptanceB, "55000",
			fixture.tenantID, &fixture.operationID},
	} {
		t.Run("usage_"+attempt.name, func(t *testing.T) {
			_, _, err := callBillingInstallationUsage(
				t, ctx, database, "billing_installation_open_usage_v2",
				fixture.operationID, attempt.tenant, attempt.manifest,
				attempt.mint,
			)
			if billingPGCode(err) != attempt.code {
				t.Fatalf("Billing usage %s code=%q err=%v",
					attempt.name, billingPGCode(err), err)
			}
			requireBillingUsageState(t, ctx, database, fixture, initial)
		})
	}
	meterCount, replay, err := callBillingInstallationUsage(
		t, ctx, database, "billing_installation_open_usage_v2",
		fixture.operationID, fixture.tenantID, fixture.manifestSHA256,
		&fixture.operationID,
	)
	if err != nil || replay || meterCount < 1 {
		t.Fatalf("Billing usage open=%d replay=%t err=%v",
			meterCount, replay, err)
	}
	opened := readBillingUsageState(
		t, ctx, database, fixture.operationID, fixture.tenantID,
	)
	if opened.openings != 1 || opened.periods != meterCount ||
		opened.totals != meterCount {
		t.Fatalf("Billing usage state=%#v meters=%d", opened, meterCount)
	}
	meterCount, replay, err = callBillingInstallationUsage(
		t, ctx, database, "billing_installation_open_usage_v2",
		fixture.operationID, fixture.tenantID, fixture.manifestSHA256,
		&fixture.operationID,
	)
	if err != nil || !replay || meterCount != opened.periods {
		t.Fatalf("Billing usage v2 replay=%d/%t err=%v",
			meterCount, replay, err)
	}
	requireBillingUsageState(t, ctx, database, fixture, opened)
}

func callBillingInstallationUsage(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	routine string,
	operationID, tenantID uuid.UUID,
	manifestSHA256 string,
	mintOperationID *uuid.UUID,
) (int, bool, error) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mintOperationID != nil {
		err = mintBillingActivationStep(
			ctx, transaction, *mintOperationID, "billing.open_tenant_usage",
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_owner`); err != nil {
		t.Fatal(err)
	}
	var count int
	var replay bool
	query := fmt.Sprintf("SELECT * FROM platform.%s($1,$2,$3)", routine)
	err = transaction.QueryRowContext(
		ctx, query, operationID, tenantID, manifestSHA256,
	).Scan(&count, &replay)
	if err != nil {
		_ = transaction.Rollback()
		_, _ = connection.ExecContext(context.Background(),
			"RESET ROLE; RESET SESSION AUTHORIZATION")
		return 0, false, err
	}
	if _, err = transaction.ExecContext(ctx,
		"RESET ROLE; RESET SESSION AUTHORIZATION"); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		return 0, false, err
	}
	return count, replay, nil
}

func readBillingUsageState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	operationID, tenantID uuid.UUID,
) billingUsageState {
	t.Helper()
	var state billingUsageState
	if err := database.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.billing_installation_usage_openings
  WHERE operation_id=$1),
 (SELECT count(*) FROM platform.billing_meter_periods WHERE tenant_id=$2),
 (SELECT count(*) FROM platform.billing_usage_totals WHERE tenant_id=$2)
`, operationID, tenantID).Scan(
		&state.openings, &state.periods, &state.totals,
	); err != nil {
		t.Fatal(err)
	}
	return state
}

func requireBillingUsageState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture billingCapabilityFixture,
	want billingUsageState,
) {
	t.Helper()
	got := readBillingUsageState(
		t, ctx, database, fixture.operationID, fixture.tenantID,
	)
	if got != want {
		t.Fatalf("Billing usage state=%#v want=%#v", got, want)
	}
}
