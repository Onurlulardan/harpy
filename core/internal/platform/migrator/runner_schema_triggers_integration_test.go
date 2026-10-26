//go:build integration

package migrator

import (
	"context"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestSchemaProgressTriggersOnPostgreSQL184(t *testing.T) {
	for _, origin := range []postgresv1.DatabaseOrigin{
		postgresv1.DatabaseOriginFresh,
		postgresv1.DatabaseOriginV1Adopted,
	} {
		t.Run(string(origin), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			database, config := startRunnerIntegrationDatabase(t, ctx)
			runner := newRunnerIntegrationFixture(t, ctx, database, config)
			if origin != postgresv1.DatabaseOriginFresh {
				prepareRunnerV1TenantRegistry(t, ctx, database)
				files, selected := runnerIntegrationCatalogForOrigin(t, origin)
				runner.platformFiles, runner.selected = files, selected
				runner.identity.databaseOrigin = origin
			}
			if _, err := runner.Up(ctx, UpRequest{RunID: runnerIntegrationRunID}); err != nil {
				t.Fatalf("apply %s schema trigger migrations: %v", origin, err)
			}
			seedSchemaTriggerRows(t, ctx, database)
			assertSchemaTransitionTriggerState(t, ctx, database)
			assertBackfillProgressAccepted(t, ctx, database)
			assertAttemptReceiptRejected(t, ctx, database, progressJobID,
				"failed", "attempt_limit_exhausted")
			assertAttemptReceiptRejected(t, ctx, database, expiredLeasedJobID,
				"lease_expired", "worker_timeout")
			assertMultiStepAttemptDivergence(t, ctx, database)
			assertActiveTerminalFailureAccepted(t, ctx, database,
				permanentFailureJobID, "ddl_permanent_failure")
			assertActiveTerminalFailureAccepted(t, ctx, database,
				activeAttemptEightJobID, "constraint_violation")
			assertSchemaRevisionMonotonicity(t, ctx, database)
			assertConcurrentSchemaRevision(t, ctx, database)
			assertExpiredExhaustionAccepted(t, ctx, database, expiredLeasedJobID)
			assertExpiredExhaustionAccepted(t, ctx, database, expiredRunningJobID)
			assertStaleSchemaTransitionsRejected(t, ctx, database)
		})
	}
}
