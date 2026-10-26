//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/domain"
)

const (
	secondTenantID = "60000000-0000-4000-8000-000000000001"
	thirdTenantID  = "60000000-0000-4000-8000-000000000002"
	fourthTenantID = "60000000-0000-4000-8000-000000000003"
)

func insertDeploymentPolicy(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	mode string,
) {
	t.Helper()
	var maximum any
	if mode == "onpremise" {
		maximum = int64(0)
	}
	_, err := database.ExecContext(ctx, `
INSERT INTO platform.tenancy_deployment_policy (
    environment_id, deployment_mode, maximum_customer_tenants
) VALUES ($1, $2, $3)`, testEnvironmentID, mode, maximum)
	if err != nil {
		t.Fatalf("insert %s deployment policy: %v", mode, err)
	}
}

func insertTestTenant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID string,
) {
	t.Helper()
	schemaName, err := domain.SchemaNameFromUUID(tenantID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `
INSERT INTO platform.tenants (
    id,name,slug,status,settings,owner_id,schema_name,created_at,updated_at
) VALUES ($1,$2,$3,'ACTIVE','{}',$4,$5,statement_timestamp(),statement_timestamp())`,
		tenantID, "Tenant "+tenantID[len(tenantID)-8:],
		"tenant-"+tenantID[len(tenantID)-8:],
		testOwnerID, schemaName.String())
	if err != nil {
		t.Fatalf("insert tenant %s: %v", tenantID, err)
	}
}

func insertTenantBinding(
	ctx context.Context,
	database *sql.DB,
	tenantID string,
	purpose string,
) error {
	_, err := database.ExecContext(ctx, `
INSERT INTO platform.tenancy_tenant_bindings (
    environment_id, tenant_id, purpose
) VALUES ($1, $2, $3)`, testEnvironmentID, tenantID, purpose)
	return err
}

func mustInsertTenantBinding(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID string,
	purpose string,
) {
	t.Helper()
	if err := insertTenantBinding(ctx, database, tenantID, purpose); err != nil {
		t.Fatalf("insert %s binding for %s: %v", purpose, tenantID, err)
	}
}

func beginRuntimeTransaction(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) *sql.Tx {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `
SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime;
SET LOCAL ROLE harpy_runtime`)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(fmt.Errorf("enter runtime authority: %w", err))
	}
	return transaction
}
