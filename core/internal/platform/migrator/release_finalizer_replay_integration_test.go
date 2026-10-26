//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

const sameSchemaReleaseID = "018f0000-0000-7000-8000-000000000504"
const sameSchemaRunID = "018f0000-0000-7000-8000-000000000505"
const sameSchemaReplayRequestID = "018f0000-0000-7000-8000-000000000506"
const sameSchemaProbeReleaseID = "018f0000-0000-7000-8000-000000000507"
const sameSchemaUnreachedProbeID = "018f0000-0000-7000-8000-000000000508"

func TestSameBinaryTupleRejectsNewReleaseAndExactReplayRecoversOnPostgreSQL184(
	t *testing.T,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	database, config := startRunnerIntegrationDatabase(t, ctx)
	createReleaseFinalizerProbe(t, ctx, database)

	initial := newRunnerIntegrationFixture(t, ctx, database, config)
	configureReleaseFinalizerIntegrationRunner(t, initial)
	initial.releaseFinalizer = &integrationReleaseFinalizer{}
	if _, err := initial.Up(
		ctx, UpRequest{RunID: successfulFinalizerRunID},
	); err != nil {
		t.Fatalf("initial release: %v", err)
	}
	assertReleaseFinalizerProbeCount(t, ctx, database, 1)
	next := newRunnerIntegrationFixture(t, ctx, database, config)
	configureReleaseFinalizerIntegrationRunner(t, next)
	configureRollingRelease(
		next,
		sameSchemaReleaseID,
		integrationDigest("migration"),
		1,
		1,
		rollingEvidence([]string{
			"018f0000-0000-7000-8000-000000000509",
			"018f0000-0000-7000-8000-000000000510",
			"018f0000-0000-7000-8000-000000000511",
		}, byte('5')),
	)
	probe := &integrationReleaseFinalizer{
		releaseID: sameSchemaProbeReleaseID, allowReady: true,
	}
	next.releaseFinalizer = probe
	_, err := next.Up(ctx, UpRequest{RunID: sameSchemaRunID})
	if err == nil ||
		!strings.Contains(err.Error(), "binary tuple is duplicated") ||
		!strings.Contains(err.Error(), "SQLSTATE 55000") ||
		!probe.called {
		t.Fatalf("duplicate binary release called=%t err=%v", probe.called, err)
	}
	assertReleaseFinalizerProbeCount(t, ctx, database, 1)

	replay := *initial
	unreached := &integrationReleaseFinalizer{
		releaseID: sameSchemaUnreachedProbeID, allowReady: true,
	}
	replay.releaseFinalizer = unreached
	recovered, err := replay.Up(
		ctx, UpRequest{RunID: sameSchemaReplayRequestID},
	)
	if err != nil || recovered.RunID != successfulFinalizerRunID ||
		len(recovered.AppliedVersions) != 0 || unreached.called {
		t.Fatalf("same-release recovery = %#v called=%t err=%v",
			recovered, unreached.called, err)
	}
	assertReleaseFinalizerProbeCount(t, ctx, database, 1)
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := postgresstore.GetMigrationRun(
		ctx, connection, sameSchemaReplayRequestID,
	); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("same-release recovery created a migration run: %v", err)
	}
}
