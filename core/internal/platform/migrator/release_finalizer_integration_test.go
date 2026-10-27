//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

const failedFinalizerRunID = "018f0000-0000-7000-8000-000000000501"
const successfulFinalizerRunID = "018f0000-0000-7000-8000-000000000502"
const finalizerProbeReleaseID = "018f0000-0000-7000-8000-000000000503"

var errFinalizerProbe = errors.New("release publication rejected")

type integrationReleaseFinalizer struct {
	releaseID  string
	allowReady bool
	fail       bool
	called     bool
}

func (finalizer *integrationReleaseFinalizer) FinalizeRelease(
	ctx context.Context,
	queryRow FinalizationQueryRow,
) error {
	finalizer.called = true
	releaseID := finalizer.releaseID
	if releaseID == "" {
		releaseID = finalizerProbeReleaseID
	}
	if finalizer.allowReady {
		var stored string
		if err := queryRow(ctx, `INSERT INTO release_finalizer_test.publications
    (release_id) VALUES ($1::uuid) RETURNING release_id::text`, releaseID).
			Scan(&stored); err != nil {
			return err
		}
		if stored != releaseID {
			return fmt.Errorf("release finalizer stored unexpected identity")
		}
		if finalizer.fail {
			return errFinalizerProbe
		}
		return nil
	}
	const statement = `WITH admission AS (
    SELECT NOT EXISTS (
        SELECT 1
        FROM platform.schema_contracts
        WHERE component = 'core' AND status = 'ready'
    ) AS allowed
),
inserted AS (
    INSERT INTO release_finalizer_test.publications (release_id)
    SELECT $1::uuid FROM admission WHERE allowed
    RETURNING release_id::text
)
SELECT release_id FROM inserted`
	var stored string
	if err := queryRow(ctx, statement, releaseID).Scan(&stored); err != nil {
		return err
	}
	if stored != releaseID {
		return fmt.Errorf("release finalizer stored unexpected identity")
	}
	if finalizer.fail {
		return errFinalizerProbe
	}
	return nil
}

func TestReleaseFinalizerSharesFoundationFinalizationTransactionOnPostgreSQL184(
	t *testing.T,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	database, config := startRunnerIntegrationDatabase(t, ctx)
	createReleaseFinalizerProbe(t, ctx, database)

	failing := newRunnerIntegrationFixture(t, ctx, database, config)
	configureReleaseFinalizerIntegrationRunner(t, failing)
	failingProbe := &integrationReleaseFinalizer{fail: true}
	failing.releaseFinalizer = failingProbe
	if _, err := failing.Up(
		ctx, UpRequest{RunID: failedFinalizerRunID},
	); !errors.Is(err, errFinalizerProbe) {
		t.Fatalf("failed finalizer result = %v", err)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	failedRun, err := postgresstore.GetMigrationRun(ctx, connection, failedFinalizerRunID)
	if err != nil || failedRun.State != "failed" ||
		failedRun.FailureCode != "release_finalization_failed" {
		t.Fatalf("failed finalizer run = %#v err=%v", failedRun, err)
	}
	assertReleaseFinalizerProbeCount(t, ctx, database, 0)
	if _, err := postgresstore.GetCoreSchemaContract(
		ctx, connection,
	); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("failed finalizer exposed ready contract: %v", err)
	}

	succeeding := newRunnerIntegrationFixture(t, ctx, database, config)
	configureReleaseFinalizerIntegrationRunner(t, succeeding)
	successProbe := &integrationReleaseFinalizer{}
	succeeding.releaseFinalizer = successProbe
	if _, err := succeeding.Up(
		ctx, UpRequest{RunID: successfulFinalizerRunID},
	); err != nil {
		t.Fatalf("successful finalizer run: %v", err)
	}
	assertReleaseFinalizerProbeCount(t, ctx, database, 1)
	contract, err := postgresstore.GetCoreSchemaContract(ctx, connection)
	if err != nil || contract.Status != "ready" ||
		contract.VerifiedRunID != successfulFinalizerRunID ||
		!failingProbe.called || !successProbe.called {
		t.Fatalf(
			"committed finalization = %#v failed_called=%t success_called=%t err=%v",
			contract, failingProbe.called, successProbe.called, err,
		)
	}
}

func configureReleaseFinalizerIntegrationRunner(t *testing.T, runner *Runner) {
	t.Helper()
	files, selected := runnerIntegrationBaselineCatalog(t)
	runner.platformFiles, runner.selected = files, selected
	runner.tenantFiles, runner.tenantSelected = nil, nil
	runner.identity.databaseGeneration = 1
	runner.identity.minimumRuntimeGeneration = 1
}

func createReleaseFinalizerProbe(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(
		ctx, `CREATE SCHEMA release_finalizer_test AUTHORIZATION harpy_owner`,
	); err != nil {
		t.Fatalf("create release finalizer probe schema: %v", err)
	}
	if _, err := database.ExecContext(ctx, `CREATE TABLE
release_finalizer_test.publications (release_id uuid PRIMARY KEY)`); err != nil {
		t.Fatalf("create release finalizer probe table: %v", err)
	}
}

func assertReleaseFinalizerProbeCount(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	want int,
) {
	t.Helper()
	var count int
	if err := database.QueryRowContext(
		ctx, `SELECT count(*) FROM release_finalizer_test.publications`,
	).Scan(&count); err != nil || count != want {
		t.Fatalf("release finalizer publication count = %d, want %d, err=%v", count, want, err)
	}
}
