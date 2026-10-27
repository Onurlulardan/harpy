//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestCatalogPG18FreshAndAdoptedLifecycleReplayIsolation(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startCatalogPG18(t, ctx, origin)
			prepareCatalogPG18(t, ctx, database, origin)
			assertCatalogMenuBaselineState(t, ctx, database.owner, "catalog_a", 1)
			assertCatalogMigrationShape(t, ctx, database)
			assertCatalogDemoOperationPlanACL(t, ctx, database)
			assertCatalogRegistryPublication(t, ctx, database.owner)
			coordinator := openCatalogCoordinator(t, ctx, database)
			assertCatalogLifecycleReplayIsolation(t, ctx, database, coordinator)
			assertCatalogDemoLifecycleAtomicity(t, ctx, database, coordinator)
			if origin == "v1_adopted" {
				assertCatalogAdoption(t, ctx, coordinator)
				assertCatalogRowCount(t, ctx, database.owner,
					"catalog_a.catalog_adoption_runs", 1)
				assertCatalogRowCount(t, ctx, database.owner,
					"catalog_b.catalog_adoption_runs", 1)
			}
		})
	}
}
