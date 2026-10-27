//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertRecoveryNonzeroActivation(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
	rootID, childID uuid.UUID,
) {
	t.Helper()
	raceCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var revision int64
	if err := scenario.db.admin.QueryRow(raceCtx, `SELECT revision
FROM t_profile_publication.iam_users WHERE id=$1`, publicationActorID).
		Scan(&revision); err != nil {
		t.Fatal(err)
	}
	worker, err := prepareRecoveryWorkerPage(raceCtx, scenario.db, childID, 1)
	if err != nil || worker.outcome.Successor != nil {
		t.Fatalf("prepare recovery activation race = %#v/%v", worker, err)
	}
	defer worker.Rollback(context.Background())
	var workerPID int
	if err = worker.transaction.QueryRow(
		raceCtx, "SELECT pg_catalog.pg_backend_pid()",
	).Scan(&workerPID); err != nil {
		t.Fatal(err)
	}
	operationID := uuid.MustParse("57000000-0000-4000-8000-000000000102")
	result := make(chan publicationRaceWriterCall, 1)
	go func() {
		written, writeErr := executePublicationRaceWriter(
			raceCtx, scenario.runtime, uuid.MustParse(scenario.db.tenant.id),
			revision, publicationNicknamePatch, operationID,
		)
		result <- publicationRaceWriterCall{result: written, err: writeErr}
	}()
	waitPublicationRaceWriterLock(t, raceCtx, scenario.db, workerPID, result)
	if err = worker.Commit(raceCtx); err != nil {
		t.Fatalf("commit recovery activation worker: %v", err)
	}
	written := <-result
	if written.err != nil || written.result.id != uuid.MustParse(publicationActorID) ||
		written.result.revision != revision+1 || written.result.replayed ||
		written.result.attempts < 1 || written.result.attempts > 2 {
		t.Fatalf("post-recovery writer = %#v/%v", written.result, written.err)
	}
	replay, err := executePublicationRaceWriter(
		raceCtx, scenario.runtime, uuid.MustParse(scenario.db.tenant.id),
		revision, publicationNicknamePatch, operationID,
	)
	if err != nil || replay.id != uuid.MustParse(publicationActorID) ||
		replay.revision != revision+1 || !replay.replayed || replay.attempts != 1 {
		t.Fatalf("post-recovery writer replay = %#v/%v", replay, err)
	}
	assertRecoveryLineageCounts(t, raceCtx, scenario.db, rootID,
		map[int64]string{0: "fenced", 1: "activated"})
	assertPublicationWorkerActivation(
		t, raceCtx, scenario.db, scenario.runtime, childID,
	)
}
