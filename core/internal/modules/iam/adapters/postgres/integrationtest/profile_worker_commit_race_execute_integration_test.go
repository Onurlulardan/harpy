//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type publicationRaceWriterCall struct {
	result publicationRaceWriterResult
	err    error
}

func assertPublicationWorkerCommitWriterRace(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime, planID uuid.UUID,
) {
	t.Helper()
	raceCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var sourceRevision int64
	if err := db.admin.QueryRow(raceCtx, `SELECT revision
FROM t_profile_publication.iam_users WHERE id=$1`, publicationActorID).
		Scan(&sourceRevision); err != nil {
		t.Fatal(err)
	}
	worker, err := preparePublicationProfileWorker(raceCtx, db, planID)
	if err != nil || worker.outcome.Successor != nil {
		t.Fatalf("prepare worker commit race = %#v/%v", worker, err)
	}
	defer worker.Rollback(context.Background())
	var workerPID int
	if err = worker.transaction.QueryRow(
		raceCtx, "SELECT pg_catalog.pg_backend_pid()",
	).Scan(&workerPID); err != nil {
		t.Fatal(err)
	}
	operationID := uuid.MustParse(
		"57000000-0000-4000-8000-000000000001",
	)
	result := make(chan publicationRaceWriterCall, 1)
	go func() {
		written, writeErr := executePublicationRaceWriter(
			raceCtx, runtime, uuid.MustParse(db.tenant.id),
			sourceRevision+1, publicationNicknamePatch, operationID,
		)
		result <- publicationRaceWriterCall{result: written, err: writeErr}
	}()
	waitPublicationRaceWriterLock(
		t, raceCtx, db, workerPID, result,
	)
	if err = worker.Commit(raceCtx); err != nil {
		t.Fatalf("commit profile worker: %v", err)
	}
	written := <-result
	if written.err != nil || written.result.id != uuid.MustParse(publicationActorID) ||
		written.result.revision != sourceRevision+2 || written.result.replayed {
		t.Fatalf("post-worker profile writer = %#v/%v", written.result, written.err)
	}
	if written.result.attempts != 2 {
		t.Fatalf("post-worker profile writer attempts = %d", written.result.attempts)
	}
	replay, err := executePublicationRaceWriter(
		raceCtx, runtime, uuid.MustParse(db.tenant.id),
		sourceRevision+1, publicationNicknamePatch, operationID,
	)
	if err != nil || replay.id != uuid.MustParse(publicationActorID) ||
		replay.revision != sourceRevision+2 || !replay.replayed || replay.attempts != 1 {
		t.Fatalf("post-worker writer replay = %#v/%v", replay, err)
	}
	assertPublicationWorkerCommitEvidence(
		t, raceCtx, db, planID, operationID, sourceRevision+2,
	)
	assertPublicationRaceTenantIsolation(
		t, raceCtx, db, runtime, sourceRevision+2,
	)
	assertPublicationWorkerActivation(t, raceCtx, db, runtime, planID)
}

func waitPublicationRaceWriterLock(
	t *testing.T, ctx context.Context, db publicationDatabase,
	workerPID int, result <-chan publicationRaceWriterCall,
) {
	t.Helper()
	observer, err := pgx.Connect(ctx, db.runtimeDSN)
	if err != nil {
		t.Fatalf("open profile race observer: %v", err)
	}
	defer func() { _ = observer.Close(context.Background()) }()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case early := <-result:
			t.Fatalf("profile writer did not wait for worker: %#v", early)
		default:
		}
		var waitingQuery string
		err = observer.QueryRow(ctx, `SELECT COALESCE((SELECT query
 FROM pg_catalog.pg_stat_activity WHERE pid<>$1
 AND application_name='harpy-v2:core-api'
 AND wait_event_type='Lock' LIMIT 1),'')`, workerPID).Scan(&waitingQuery)
		if err != nil {
			t.Fatalf("profile writer wait state: %v", err)
		}
		if waitingQuery != "" {
			t.Logf("profile writer lock query: %s", waitingQuery)
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("profile writer did not block: %v", ctx.Err())
		case <-ticker.C:
		}
	}
}
