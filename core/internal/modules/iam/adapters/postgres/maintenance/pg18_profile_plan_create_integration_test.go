//go:build integration

package maintenance

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func createMaintenancePG18ProfilePlan(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant maintenancePG18Tenant, planID uuid.UUID, document []byte,
	planSHA, requestSHA [32]byte, tenantIndex int,
) {
	t.Helper()
	createMaintenancePG18ProfilePlanWithOperation(
		t, ctx, database, tenant, planID, document, planSHA, requestSHA,
		uuid.MustParse([]string{
			"48000000-0000-4000-8000-000000000001",
			"48000000-0000-4000-8000-000000000002",
		}[tenantIndex]), true,
	)
}

func createMaintenancePG18ProfilePlanWithOperation(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant maintenancePG18Tenant, planID uuid.UUID, document []byte,
	planSHA, requestSHA [32]byte, operationID uuid.UUID, verifyReplay bool,
) {
	t.Helper()
	attempts := 1
	if verifyReplay {
		attempts = 2
	}
	for attempt := 0; attempt < attempts; attempt++ {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, statement := range []string{
			"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
			"SET LOCAL ROLE harpy_runtime",
		} {
			if _, err = transaction.ExecContext(ctx, statement); err != nil {
				break
			}
		}
		if err == nil {
			_, err = transaction.ExecContext(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)",
				tenant.schema+", platform, pg_catalog")
		}
		var resultID uuid.UUID
		var revision int64
		var replayed bool
		if err == nil {
			err = transaction.QueryRowContext(ctx, `SELECT * FROM
iam_profile_backfill_plan_create_v2(
 pg_catalog.current_schema(),$1::uuid,$2::uuid,$3::bytea,$4::bytea,
 $5::bytea,$6::uuid,$7::uuid,$8::bigint,$9::bigint)`,
				tenant.id, operationID, requestSHA[:], document, planSHA[:],
				maintenanceUserThree, maintenanceReadPermission, int64(1), int64(1),
			).Scan(&resultID, &revision, &replayed)
		}
		if err == nil {
			err = transaction.Commit()
		} else {
			_ = transaction.Rollback()
		}
		if err != nil || resultID != planID || revision != 1 ||
			replayed != (verifyReplay && attempt == 1) {
			t.Fatalf("%s create profile plan=%s/%d/%t/%x: %v",
				tenant.schema, resultID, revision, replayed,
				sha256.Sum256(document), err)
		}
	}
}
