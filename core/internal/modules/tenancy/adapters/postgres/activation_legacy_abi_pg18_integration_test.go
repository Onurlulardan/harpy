//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestActivationLegacyABICompatibleAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startTenantMigrationDatabase(t, ctx)
			prepareActivationScopeDatabaseThrough(
				t, ctx, database, origin, 20260729100036,
			)
			assertLegacyActivationPrivilegesRetained(t, ctx, database)

			fixture := newActivationScopeFixture(t,
				"environment_initialize", "saas_control", "legacy-v3-"+origin)
			insertActivationOperation(t, ctx, database, fixture)
			callLegacyActivationABI(t, ctx, database, fixture, 3)
			var tenants, facts int
			err := database.QueryRowContext(ctx, `SELECT
(SELECT pg_catalog.count(*) FROM platform.tenants WHERE id=$1),
(SELECT pg_catalog.count(*) FROM platform.tenancy_activation_facts
 WHERE tenant_id=$1)`, fixture.tenantID).Scan(&tenants, &facts)
			if err != nil {
				t.Fatal(err)
			}
			if tenants != 1 || facts != 1 {
				t.Fatalf("legacy activation v3 tenant/facts = %d/%d, want 1/1",
					tenants, facts)
			}
		})
	}
}

func assertLegacyActivationPrivilegesRetained(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for version := 1; version <= 3; version++ {
		signature := fmt.Sprintf(
			"platform.tenancy_activation_initialize_v%d"+
				"(uuid,uuid,uuid,text,text,uuid,text,text,text)",
			version,
		)
		var allowed bool
		if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_function_privilege(
    'harpy_runtime',$1,'EXECUTE'
)`, signature).Scan(&allowed); err != nil {
			t.Fatal(err)
		}
		if !allowed {
			t.Fatalf("harpy_runtime lost activation v%d EXECUTE", version)
		}
	}
}

func callLegacyActivationABI(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture activationScopeFixture, version int,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `
SELECT pg_catalog.set_config(
    'application_name','harpy-v2:core-operator-initialize',true
)`); err != nil {
		t.Fatal(err)
	}
	statement := fmt.Sprintf(`
SELECT platform.tenancy_activation_initialize_v%d(
    $1,$2,$3,$4,$5,$6,$7,$8,$9
	)`, version)
	var replayed bool
	if err := transaction.QueryRowContext(ctx, statement,
		fixture.operationID, fixture.environmentID, fixture.tenantID,
		fixture.name, fixture.slug, fixture.ownerID, fixture.schema,
		fixture.purpose, fixture.manifest,
	).Scan(&replayed); err != nil {
		t.Fatalf("activation v%d failed: %v", version, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("activation v%d commit failed: %v", version, err)
	}
}
