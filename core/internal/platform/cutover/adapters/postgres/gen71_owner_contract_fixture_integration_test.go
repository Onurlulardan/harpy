//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
)

func applyCutoverMigrationPrerequisites(
	t *testing.T, ctx context.Context, connection *pgx.Conn,
	origin, migrationName string,
) {
	t.Helper()
	gen71Fence := "20260729100045_assert_gen71_activation_contract_v1." +
		origin + ".sql"
	if migrationName == gen71Fence {
		applyGen71OwnerContracts(t, ctx, connection, origin)
	}
}

func applyGen71OwnerContracts(
	t *testing.T, ctx context.Context, connection *pgx.Conn, origin string,
) {
	t.Helper()
	contracts := []string{
		"../../../../modules/tenancy/migrations/platform/" + origin +
			"/20260729100037_contract_gen71_tenancy_legacy_writers_v1." + origin + ".sql",
		"../../../../modules/iam/migrations/platform/" + origin +
			"/20260729100038_contract_iam_installation_platform_authority_v1." + origin + ".sql",
		"../../../../modules/billing/migrations/platform/" + origin +
			"/20260729100039_contract_gen71_billing_legacy_writers_v1." + origin + ".sql",
		"../../../../modules/analytics/migrations/platform/" + origin +
			"/20260729100040_contract_analytics_activation_authority_v1." + origin + ".sql",
		"../../../../modules/workflow/migrations/platform/" + origin +
			"/20260729100041_contract_workflow_activation_authority_v1." + origin + ".sql",
		"../../../../modules/platformops/migrations/platform/" + origin +
			"/20260729100042_contract_installation_recovery_authority_v1." + origin + ".sql",
		"../../../async/migrations/" + origin +
			"/20260729100043_contract_installation_activation_writers_v1." + origin + ".sql",
		"../../../../modules/platformauthority/migrations/platform/" + origin +
			"/20260729100044_contract_installation_authority_pins_v1." + origin + ".sql",
	}
	for _, path := range contracts {
		document, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = connection.Exec(ctx, string(document)); err != nil {
			t.Fatalf("apply Gen71 owner contract %s: %v", path, err)
		}
	}
}
