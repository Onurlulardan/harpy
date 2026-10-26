//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestBillingAuthorityLifecycleAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(), 5*time.Minute,
			)
			defer cancel()
			database := prepareBillingOrigin(t, ctx, origin)

			assertBillingAuthorityReconcile(t, ctx, database)
			assertBillingGen71Contract(t, ctx, database)
			assertBillingInstallationUsesCurrentGeneration(t, ctx, database)
			assertBillingInstallationUsageCapability(t, ctx, database)
			assertBillingInstallationCapabilityNegatives(t, ctx, database)
			assertBillingAuthorityReconcileRace(t, ctx, database)
			assertBillingFinalizedAuthoritySuccessors(t, ctx, database)
		})
	}
}

func TestBillingGeneration23CompatibilityAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(), 5*time.Minute,
			)
			defer cancel()
			database := prepareBillingOriginThrough(
				t, ctx, origin, 20260729100036,
			)
			assertBillingExpandCompatibility(t, ctx, database)
		})
	}
}
