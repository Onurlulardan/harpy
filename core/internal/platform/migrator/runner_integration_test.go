//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func TestRunnerPlanVerifyUpAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database, config := startRunnerIntegrationDatabase(t, ctx)
	runner := newRunnerIntegrationFixture(t, ctx, database, config)
	baselineFiles, baselineSelected := runnerIntegrationBaselineCatalog(t)
	runner.platformFiles, runner.selected = baselineFiles, baselineSelected
	runner.tenantFiles, runner.tenantSelected = nil, nil
	runner.identity.databaseGeneration = 1
	plan, err := runner.Plan(ctx)
	if err != nil {
		t.Fatalf("plan fresh foundation: %v", err)
	}
	if plan.CurrentVersion != 0 || len(plan.Pending) != len(runner.selected) {
		t.Fatalf("fresh plan = version %d pending %d", plan.CurrentVersion, len(plan.Pending))
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("open assertion connection: %v", err)
	}
	defer connection.Close()
	if _, exists, err := postgresstore.ReadAppliedHistory(ctx, connection); err != nil || exists {
		t.Fatalf("plan created migration history: exists=%t err=%v", exists, err)
	}
	if _, err := runner.Verify(ctx); !errors.Is(err, ErrMigrationsPending) {
		t.Fatalf("fresh verify error = %v, want pending", err)
	}
	result, err := runner.Up(ctx, UpRequest{RunID: runnerIntegrationRunID})
	if err != nil {
		t.Fatalf("apply fresh foundation: %v", err)
	}
	lastVersion := runner.selected[len(runner.selected)-1].Version
	if result.DatabaseVersion != lastVersion ||
		len(result.AppliedVersions) != len(runner.selected) {
		t.Fatalf("up result = version %d applied %d",
			result.DatabaseVersion, len(result.AppliedVersions))
	}
	verified, err := runner.Verify(ctx)
	if err != nil {
		t.Fatalf("verify migrated foundation: %v", err)
	}
	if verified.DatabaseVersion != lastVersion || verified.DatabaseGeneration != 1 ||
		verified.MinimumRuntimeGeneration != 1 ||
		verified.VerifiedRunID != runnerIntegrationRunID {
		t.Fatalf("verified foundation = %#v", verified)
	}
	run, err := postgresstore.GetMigrationRun(ctx, connection, runnerIntegrationRunID)
	if err != nil || run.State != "succeeded" {
		t.Fatalf("migration run = %#v err=%v", run, err)
	}
	ordered, err := postgresstore.FoundationFinalizationOrderVerified(ctx, connection)
	if err != nil || !ordered {
		t.Fatalf("foundation ready was not the final write: ordered=%t err=%v", ordered, err)
	}
	defaultHistory, err := postgresstore.DefaultGooseHistoryExists(ctx, connection)
	if err != nil || defaultHistory {
		t.Fatalf("default Goose history exists=%t err=%v", defaultHistory, err)
	}
	finalPlan, err := runner.Plan(ctx)
	if err != nil || len(finalPlan.Pending) != 0 || finalPlan.CurrentVersion != lastVersion {
		t.Fatalf("final plan = %#v err=%v", finalPlan, err)
	}
	replayed, err := runner.Up(ctx, UpRequest{RunID: runnerIntegrationRunID})
	if err != nil || len(replayed.AppliedVersions) != 0 {
		t.Fatalf("idempotent up = %#v err=%v", replayed, err)
	}
	assertRunnerLockCancellation(t, database, runner)
}

func assertRunnerLockCancellation(t *testing.T, database *sql.DB, runner *Runner) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("open lock blocker: %v", err)
	}
	defer connection.Close()
	if err := postgresstore.AcquirePlatformLock(ctx, connection, platformMigrationLockKey); err != nil {
		t.Fatalf("acquire lock blocker: %v", err)
	}
	blockedContext, cancelBlocked := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancelBlocked()
	if _, err := runner.Plan(blockedContext); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("blocked plan error = %v, want deadline", err)
	}
	if err := postgresstore.ReleasePlatformLock(ctx, connection, platformMigrationLockKey); err != nil {
		t.Fatalf("release lock blocker: %v", err)
	}
}
