//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func insertBillingShadowTenant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) uuid.UUID {
	t.Helper()
	tenantID := billingV7(t)
	schema := "t_" + strings.ReplaceAll(tenantID.String()[:12], "-", "")
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
INSERT INTO platform.tenants(id,schema_name,status,deleted_at)
VALUES($1,$2,'ACTIVE',NULL)`,
		tenantID, schema); err != nil {
		t.Fatalf("insert Billing shadow tenant: %v", err)
	}
	fixture := newBillingCutoverAuthorityFixture(
		"tenant", tenantID, "billing_forms", 1,
		"v1", "v1", billingAcceptanceA, "billing", "forms",
	)
	if err = seedBillingCutoverAuthority(ctx, transaction, fixture); err != nil {
		t.Fatalf("insert Billing shadow authority: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return tenantID
}

func assertBillingShadowTenantIgnored(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID uuid.UUID,
) {
	t.Helper()
	var count int
	if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM platform.billing_tenant_authorities
WHERE tenant_id=$1`, tenantID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("shadow tenant authority count=%d, want 0", count)
	}
}
