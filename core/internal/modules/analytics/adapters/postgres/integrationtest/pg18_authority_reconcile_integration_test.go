//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestAnalyticsAuthorityReconcileAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "authority-"+origin)
			prepareAnalyticsPrerequisites(t, ctx, database, origin)
			applyAnalyticsDocuments(t, ctx, database, "",
				analyticsDocuments(t, origin, "platform"))

			reconciled, remaining := reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 2 || remaining != 0 {
				t.Fatalf("initial reconcile = %d/%d, want 2/0",
					reconciled, remaining)
			}
			reconciled, remaining = reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 0 || remaining != 0 {
				t.Fatalf("replay reconcile = %d/%d, want 0/0",
					reconciled, remaining)
			}
			if _, err := database.ExecContext(ctx, `
UPDATE platform.cutover_groups SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='analytics'`,
				analyticsTenantA); err != nil {
				t.Fatal(err)
			}
			if _, err := database.ExecContext(ctx, `
UPDATE platform.domain_cutovers SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND domain='analytics'`,
				analyticsTenantA); err != nil {
				t.Fatal(err)
			}
			reconciled, remaining = reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 1 || remaining != 0 {
				t.Fatalf("stale-generation reconcile = %d/%d, want 1/0",
					reconciled, remaining)
			}
			assertAnalyticsAuthorityGenerations(t, ctx, database)
			reconcileAnalyticsAuthorities(t, ctx, database, false)
			assertAnalyticsAuthorityReconcileDeniedToOwner(
				t, ctx, database,
			)
		})
	}
}

func assertAnalyticsAuthorityReconcileDeniedToOwner(
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
	if _, err = connection.ExecContext(ctx,
		`SELECT pg_catalog.set_config('application_name',$1,false)`,
		"harpy-v2:core-operator-initialize"); err != nil {
		t.Fatal(err)
	}
	var reconciled, remaining int
	err = connection.QueryRowContext(ctx, `
SELECT reconciled_count,remaining_count
FROM platform.analytics_reconcile_current_tenant_authorities_v2()`,
	).Scan(&reconciled, &remaining)
	if err == nil {
		t.Fatal("Analytics authority reconciliation accepted owner session")
	}
	requireAnalyticsPGCode(t, err, "42501", "authority reconcile session user")
}

func assertAnalyticsAuthorityGenerations(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var total, current int
	if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*),
       pg_catalog.count(*) FILTER (
           WHERE authority_group_generation=2
             AND authority_domain_generation=2
       )
FROM platform.analytics_tenant_authorities
WHERE tenant_id=$1`, analyticsTenantA).Scan(&total, &current); err != nil {
		t.Fatal(err)
	}
	if total != 2 || current != 1 {
		t.Fatalf("Analytics authority generations = %d/%d, want 2/1",
			total, current)
	}
}
