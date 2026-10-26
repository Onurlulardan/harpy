//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMAuthRuntimeBehaviorAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMPlatformGuardPG18(t, ctx, "auth_behavior_"+origin)
			prepareIAMPlatformGuardSchema(t, ctx, database, origin)
			prepareIAMAuthTenantRuntime(t, ctx, database, origin)
			fixtures := seedIAMAuthRuntimeFixtures(t, ctx, database)
			applyIAMLifecycleGuards(t, ctx, database.migrator, origin)
			applyIAMAtomicRoutines(t, ctx, database.migrator, origin)
			applyIAMAuthRuntime(t, ctx, database.migrator, origin)
			grantIAMAuthTenantRuntime(t, ctx, database)

			for _, fixture := range fixtures {
				assertIAMAuthLoginAndSession(t, ctx, database, fixture)
				createIAMExpiredFixtureLocator(t, ctx, database, fixture)
				assertIAMAuthIdleExpiry(t, ctx, database, fixture)
			}
			assertIAMAuthTenantIsolation(t, ctx, database, fixtures)
			for _, fixture := range fixtures {
				assertIAMAuthRefreshConcurrency(t, ctx, database, fixture)
			}
		})
	}
}

func prepareIAMAuthTenantRuntime(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
	origin string,
) {
	t.Helper()
	if _, err := database.migrator.ExecContext(ctx, `
CREATE SCHEMA t_iam_a AUTHORIZATION harpy_owner;
CREATE SCHEMA t_iam_b AUTHORIZATION harpy_owner;`); err != nil {
		t.Fatalf("create IAM auth tenant schemas: %v", err)
	}
	documents := loadIAMFoundationDocuments(t, origin, "tenant")
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(
			t, ctx, database.migrator, tenant.schema, documents,
		)
	}
	seedIAMFoundationBase(t, ctx, database.migrator)
}

func grantIAMAuthTenantRuntime(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
) {
	t.Helper()
	if _, err := database.migrator.ExecContext(ctx, `
GRANT USAGE ON SCHEMA t_iam_a TO harpy_runtime;
GRANT USAGE ON SCHEMA t_iam_b TO harpy_runtime;`); err != nil {
		t.Fatalf("grant IAM auth tenant runtime usage: %v", err)
	}
}
