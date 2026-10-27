//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func assertBillingExpandCompatibility(
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
		"platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,bigint,timestamptz)",
	} {
		var allowed bool
		if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_function_privilege(
 'harpy_runtime',$1,'EXECUTE'
)`, signature).Scan(&allowed); err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Fatalf("generation-23 Billing ABI was revoked: %s", signature)
		}
	}
	for signature, want := range map[string]bool{
		"platform.billing_installation_assign_v3(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)": true,
		"platform.billing_reconcile_current_global_authority_v2()":                                                                  false,
		"platform.billing_reconcile_current_tenant_authorities_v2()":                                                                false,
	} {
		var allowed bool
		if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_function_privilege(
 'harpy_runtime',$1,'EXECUTE'
)`, signature).Scan(&allowed); err != nil {
			t.Fatal(err)
		}
		if allowed != want {
			t.Fatalf("Billing successor EXECUTE=%t want=%t: %s",
				allowed, want, signature)
		}
	}
	tenantID := billingV7(t)
	insertBillingTenantTopology(
		t, ctx, database, tenantID, 4, billingAcceptanceA,
	)
	offering := loadBillingControlOffering(t, ctx, database)
	operationID := billingV7(t)
	authorityID := billingInstallationIDFixture(
		operationID, "tenant-authority",
	)
	assignmentID := billingInstallationIDFixture(
		operationID, "base-offering-assignment",
	)
	replayed, err := callGeneration23BillingAssign(
		t, ctx, database, operationID, tenantID,
		authorityID, assignmentID, offering,
	)
	if err != nil || replayed {
		t.Fatalf("generation-23 Billing assign replay=%t err=%v", replayed, err)
	}
	assertGeneration23BillingReconcile(t, ctx, database)
	before := countBillingInstallationRows(t, ctx, database, operationID)
	_, err = callGeneration23BillingAssign(
		t, ctx, database, operationID, tenantID,
		billingV7(t), assignmentID, offering,
	)
	if billingPGCode(err) != "23505" {
		t.Fatalf("mismatched generation-23 Billing code=%q err=%v",
			billingPGCode(err), err)
	}
	after := countBillingInstallationRows(t, ctx, database, operationID)
	if before != 1 || after != before {
		t.Fatalf("mismatched generation-23 Billing wrote rows: %d/%d",
			before, after)
	}
}

func callGeneration23BillingAssign(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	operationID, tenantID, authorityID, assignmentID uuid.UUID,
	offering billingInstallationOffering,
) (bool, error) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(),
			"RESET ROLE; RESET SESSION AUTHORIZATION")
	}()
	var replayed bool
	err = connection.QueryRowContext(ctx, `
SELECT platform.billing_installation_assign_v2(
 $1,$2,$3,$4,$5,'saas',$6,$7,$8,$9,NULL,NULL,$10,'saas_control'
)`, operationID, uuid.MustParse(billingEnvironmentID), tenantID,
		authorityID, assignmentID, offering.versionID, offering.revision,
		offering.sha256, offering.effective, billingAcceptanceA,
	).Scan(&replayed)
	return replayed, err
}

func assertGeneration23BillingReconcile(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_runtime;
SET application_name='harpy-v2:core-operator-initialize'`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(),
			"RESET ROLE; RESET SESSION AUTHORIZATION")
	}()
	for _, routine := range []string{
		"billing_reconcile_current_global_authority_v1",
		"billing_reconcile_current_tenant_authorities_v1",
	} {
		var reconciled, remaining int
		err = connection.QueryRowContext(ctx,
			"SELECT * FROM platform."+routine+"()",
		).Scan(&reconciled, &remaining)
		if err != nil || remaining != 0 {
			t.Fatalf("generation-23 %s=%d/%d err=%v",
				routine, reconciled, remaining, err)
		}
	}
}

func countBillingInstallationRows(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	operationID uuid.UUID,
) int {
	t.Helper()
	var count int
	if err := database.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_installation_base_assignments
WHERE operation_id=$1`, operationID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
