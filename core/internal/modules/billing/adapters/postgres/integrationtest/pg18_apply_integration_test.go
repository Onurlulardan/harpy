//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestBillingMigrationsAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startBillingPG18(t, ctx, origin)
			prepareBillingPrerequisites(t, ctx, database, origin)
			applyBillingDocuments(t, ctx, database, loadBillingDocuments(t, origin))
			assertBillingCatalog(t, ctx, database)
			assertBillingV1RetirementInventory(t, ctx, database)
			assertBillingLicenseLifecycle(t, ctx, database)
		})
	}
}
