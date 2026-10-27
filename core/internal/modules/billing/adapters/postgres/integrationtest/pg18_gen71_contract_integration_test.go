//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func assertBillingGen71Contract(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, signature := range []string{
		"platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text)",
		"platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)",
		"platform.billing_installation_open_usage_v1(uuid,uuid,text)",
		"platform.billing_reconcile_current_global_authority_v1()",
		"platform.billing_reconcile_current_tenant_authorities_v1()",
		"platform.billing_register_global_authority_v1(uuid,bigint,bigint,timestamptz)",
		"platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,timestamptz)",
	} {
		requireBillingGen71FunctionACL(t, ctx, database, signature, false)
	}
	for _, signature := range []string{
		"platform.billing_installation_assign_v3(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)",
		"platform.billing_installation_open_usage_v2(uuid,uuid,text)",
	} {
		requireBillingGen71FunctionACL(t, ctx, database, signature, true)
	}
	for _, signature := range []string{
		"platform.billing_installation_id_v1(uuid,text)",
		"platform.billing_reconcile_current_global_authority_v2()",
		"platform.billing_reconcile_current_tenant_authorities_v2()",
	} {
		requireBillingGen71FunctionACL(t, ctx, database, signature, false)
	}

	tenantID := billingV7(t)
	offering := loadBillingControlOffering(t, ctx, database)
	operationID := billingV7(t)
	_, err := callGeneration23BillingAssign(
		t, ctx, database, operationID, tenantID,
		billingInstallationIDFixture(operationID, "tenant-authority"),
		billingInstallationIDFixture(operationID, "base-offering-assignment"),
		offering,
	)
	if billingPGCode(err) != "42501" {
		t.Fatalf("retired Billing assign code=%q err=%v",
			billingPGCode(err), err)
	}
	if rows := countBillingInstallationRows(
		t, ctx, database, operationID,
	); rows != 0 {
		t.Fatalf("retired Billing assign wrote %d rows", rows)
	}

	fixture := billingCapabilityFixture{
		operationID: billingV7(t),
		tenantID:    billingV7(t),
	}
	before := readBillingUsageState(
		t, ctx, database, fixture.operationID, fixture.tenantID,
	)
	err = callRetiredBillingUsage(
		t, ctx, database, fixture.operationID, fixture.tenantID,
		billingAcceptanceA,
	)
	if billingPGCode(err) != "42501" {
		t.Fatalf("retired Billing usage code=%q err=%v",
			billingPGCode(err), err)
	}
	requireBillingUsageState(t, ctx, database, fixture, before)
}

func callRetiredBillingUsage(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	operationID, tenantID uuid.UUID,
	manifestSHA256 string,
) error {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime;
SET LOCAL ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	var count int
	var replay bool
	return transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_installation_open_usage_v1($1,$2,$3)`,
		operationID, tenantID, manifestSHA256,
	).Scan(&count, &replay)
}

func requireBillingGen71FunctionACL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	signature string,
	runtimeAllowed bool,
) {
	t.Helper()
	var runtime, login bool
	err := database.QueryRowContext(ctx, `SELECT
pg_catalog.has_function_privilege('harpy_runtime',$1,'EXECUTE'),
pg_catalog.has_function_privilege('harpy_v2_runtime',$1,'EXECUTE')`,
		signature).Scan(&runtime, &login)
	if err != nil || runtime != runtimeAllowed || login {
		t.Fatalf("Billing ACL %s=%t/%t want=%t/false err=%v",
			signature, runtime, login, runtimeAllowed, err)
	}
}
