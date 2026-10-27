//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
	"time"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

const (
	iamMaintenanceJobID         = "019c6a7e-7c00-7000-8000-000000000010"
	iamMaintenanceKeyID         = "AAAAAAAAAAAAAAAA"
	iamMaintenanceSourceID      = "61000000-0000-4000-8000-000000000001"
	iamMaintenanceRepairJobB    = "019c6a7e-7c00-7000-8000-000000000020"
	iamMaintenanceChecksumJobA  = "019c6a7e-7c00-7000-8000-000000000030"
	iamMaintenanceChecksumJobB  = "019c6a7e-7c00-7000-8000-000000000040"
	iamMaintenanceAttemptA      = "71000000-0000-4000-8000-000000000001"
	iamMaintenanceAttemptB      = "71000000-0000-4000-8000-000000000002"
	iamMaintenanceAttemptResume = "71000000-0000-4000-8000-000000000003"
	iamMaintenanceAttemptCheckA = "71000000-0000-4000-8000-000000000004"
	iamMaintenanceAttemptCheckB = "71000000-0000-4000-8000-000000000005"
	iamMaintenanceNextRepairA   = "019c6a7e-7c00-7000-8000-000000000110"
	iamMaintenanceNextRepairB   = "019c6a7e-7c00-7000-8000-000000000120"
	iamMaintenanceNextChecksumA = "019c6a7e-7c00-7000-8000-000000000130"
	iamMaintenanceNextChecksumB = "019c6a7e-7c00-7000-8000-000000000140"
)

func seedMaintenancePG18Directory(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `
INSERT INTO iam_user_identifiers(
 id,user_id,kind,display_value,normalized_value,normalized_sha256,
 state,revision,verified_at,activated_at,tombstoned_at,cooldown_until,
 retired_at,created_at,updated_at)
VALUES($1::uuid,$2::uuid,'email_v1',$3,$3,
 pg_catalog.sha256(pg_catalog.convert_to($3,'UTF8')),'active',1,
 '2026-07-16T01:00:00Z','2026-07-16T01:00:00Z',NULL,NULL,NULL,
 '2026-07-16T01:00:00Z','2026-07-16T01:00:00Z')`
	for index, tenant := range maintenancePG18Tenants {
		normalized := []string{"repair-a@example.test", "repair-b@example.test"}[index]
		execMaintenancePG18Tenant(t, ctx, database, tenant, statement,
			iamMaintenanceSourceID, maintenanceUserOne, normalized)
	}
}

func bootstrapMaintenancePG18Schedule(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant maintenancePG18Tenant, kind maintenanceapp.Kind, jobID string,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	_, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		tenant.schema+", platform, pg_catalog")
	var reference string
	var digest []byte
	var created bool
	if err == nil {
		err = transaction.QueryRowContext(ctx, `SELECT * FROM
iam_maintenance_bootstrap_v1(pg_catalog.current_schema(),$1::uuid,$2,$3::uuid,$4)`,
			tenant.id, string(kind), jobID,
			time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC),
		).Scan(&reference, &digest, &created)
	}
	if err != nil || reference == "" || len(digest) != 32 {
		t.Fatalf("bootstrap %s/%s=%q/%d/%t: %v",
			tenant.schema, kind, reference, len(digest), created, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func nextMaintenancePG18Claim(
	t *testing.T, tenantID string, kind maintenanceapp.Kind,
	jobID uuid.UUID, attempt string,
) maintenanceapp.Claim {
	return recurringMaintenancePG18Claim(t, tenantID, jobID.String(), attempt, kind)
}
