//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
)

func prepareLifecycleOrigin(t *testing.T, ctx context.Context,
	database *sql.DB, origin string,
) {
	t.Helper()
	if origin == "v1_adopted" {
		prepareV1TenantRegistry(t, ctx, database)
	}
	runPlatformMigrations(t, ctx, database, origin)
	insertTestTenant(t, ctx, database, lifecycleControlID)
	insertTestTenant(t, ctx, database, lifecycleTargetID)
	insertTestTenant(t, ctx, database, lifecyclePurgeID)
	insertDeploymentPolicy(t, ctx, database, "saas")
	mustInsertTenantBinding(t, ctx, database, lifecycleControlID, "saas_control")
	mustInsertTenantBinding(t, ctx, database, lifecycleTargetID, "saas_customer")
	mustInsertTenantBinding(t, ctx, database, lifecyclePurgeID, "saas_customer")
}

func assertLifecyclePrivileges(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var tenantsRead, receiptsWrite, operationsWrite, listExecute, purgeExecute bool
	err := database.QueryRowContext(ctx, `SELECT
has_table_privilege('harpy_runtime','platform.tenants','SELECT'),
has_table_privilege('harpy_runtime','platform.tenancy_customer_lifecycle_receipts','INSERT'),
has_table_privilege('harpy_runtime','platform.tenancy_customer_lifecycle_operations','UPDATE'),
has_function_privilege('harpy_runtime','platform.tenancy_customer_lifecycle_list_v1(uuid,integer,uuid)','EXECUTE'),
has_function_privilege('harpy_runtime','platform.tenancy_customer_lifecycle_complete_purge_v1(uuid,uuid,uuid,bytea,timestamptz)','EXECUTE')
`).Scan(&tenantsRead, &receiptsWrite, &operationsWrite, &listExecute, &purgeExecute)
	if err != nil || tenantsRead || receiptsWrite || operationsWrite ||
		!listExecute || !purgeExecute {
		t.Fatalf("lifecycle privileges tenant=%t receipts=%t operations=%t list=%t purge=%t err=%v",
			tenantsRead, receiptsWrite, operationsWrite, listExecute, purgeExecute, err)
	}
}

func assertLifecycleReads(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	var document []byte
	var next sql.NullString
	err := transaction.QueryRowContext(ctx, `SELECT document,next_cursor
FROM platform.tenancy_customer_lifecycle_list_v1($1,1,NULL)`, testEnvironmentID,
	).Scan(&document, &next)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	var values []map[string]any
	if json.Unmarshal(document, &values) != nil || len(values) != 1 || !next.Valid {
		_ = transaction.Rollback()
		t.Fatalf("lifecycle page=%s next=%+v", document, next)
	}
	if values[0]["id"] == lifecycleControlID {
		_ = transaction.Rollback()
		t.Fatal("control tenant leaked into customer list")
	}
	var id, name, slug, status string
	var revision, settingsRevision int64
	var createdAt, updatedAt any
	var offboardedAt, purgedAt any
	err = transaction.QueryRowContext(ctx, `SELECT *
FROM platform.tenancy_customer_lifecycle_get_v1($1,$2)`,
		testEnvironmentID, lifecycleTargetID).Scan(&id, &name, &slug, &status,
		&revision, &settingsRevision, &createdAt, &updatedAt, &offboardedAt, &purgedAt)
	if err != nil || id != lifecycleTargetID || revision != 1 || settingsRevision != 1 {
		_ = transaction.Rollback()
		t.Fatalf("lifecycle get id=%s revision=%d/%d err=%v", id, revision, settingsRevision, err)
	}
	var controlCount int
	err = transaction.QueryRowContext(ctx, `SELECT pg_catalog.count(*)
FROM platform.tenancy_customer_lifecycle_get_v1($1,$2)`,
		testEnvironmentID, lifecycleControlID).Scan(&controlCount)
	if err != nil || controlCount != 0 {
		_ = transaction.Rollback()
		t.Fatalf("control customer read count=%d err=%v", controlCount, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertLifecycleReceiptsImmutable(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `UPDATE platform.tenancy_customer_lifecycle_receipts
SET result_revision=result_revision+1 WHERE tenant_id=$1`, lifecycleTargetID); err == nil {
		t.Fatal("customer lifecycle receipt mutation was accepted")
	}
	if _, err := database.ExecContext(ctx, `UPDATE platform.tenancy_customer_lifecycle_operations
SET result_revision=result_revision+1 WHERE state='SUCCEEDED'`); err == nil {
		t.Fatal("terminal customer lifecycle operation mutation was accepted")
	}
	if _, err := database.ExecContext(ctx, `DELETE FROM platform.tenancy_customer_lifecycle_operations
WHERE state='SUCCEEDED'`); err == nil {
		t.Fatal("customer lifecycle operation deletion was accepted")
	}
}
