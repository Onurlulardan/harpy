//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

type profileRaceCallResult struct {
	page     int
	revision int64
	err      error
}

func assertMaintenancePG18ProfileLiveRaces(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	assertMaintenancePG18LiveWinsProfileRace(
		t, ctx, database, fixture, tenantIndex,
	)
	assertMaintenancePG18WorkerRollbackProfileRace(
		t, ctx, database, fixture, tenantIndex,
	)
	assertMaintenancePG18ProfileRaceStored(
		t, ctx, database, fixture, tenantIndex,
	)
}

func assertMaintenancePG18LiveWinsProfileRace(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	raceCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	live := beginProfileRuntimeTransaction(
		t, raceCtx, database, fixture.tenant.schema,
	)
	defer live.Rollback()
	page, revision, err := executeMaintenancePG18RaceLiveWrite(
		raceCtx, live, fixture, tenantIndex, 1,
	)
	if err != nil || page != 2 || revision != 5 {
		t.Fatalf("%s live-first prepare=%d/%d: %v",
			fixture.tenant.schema, page, revision, err)
	}
	worker := beginProfileRuntimeTransaction(
		t, raceCtx, database, fixture.tenant.schema,
	)
	defer worker.Rollback()
	workerPID := maintenancePG18ProfileRacePID(t, raceCtx, worker)
	result := make(chan profileRaceCallResult, 1)
	go func() {
		workerRevision, workerErr := executeMaintenancePG18RaceWorker(
			raceCtx, worker, fixture, tenantIndex, 1,
		)
		result <- profileRaceCallResult{revision: workerRevision, err: workerErr}
	}()
	waitMaintenancePG18ProfileRaceLock(
		t, raceCtx, database, workerPID, "worker after live",
	)
	if err = live.Commit(); err != nil {
		t.Fatalf("%s live-first commit: %v", fixture.tenant.schema, err)
	}
	workerResult := <-result
	if workerResult.err == nil || !strings.Contains(
		strings.ToLower(workerResult.err.Error()), "source changed",
	) {
		t.Fatalf("%s live-first worker=%d: %v",
			fixture.tenant.schema, workerResult.revision, workerResult.err)
	}
}

func assertMaintenancePG18WorkerRollbackProfileRace(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture maintenancePlanFixture, tenantIndex int,
) {
	t.Helper()
	raceCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	worker := beginProfileRuntimeTransaction(
		t, raceCtx, database, fixture.tenant.schema,
	)
	defer worker.Rollback()
	revision, err := executeMaintenancePG18RaceWorker(
		raceCtx, worker, fixture, tenantIndex, 2,
	)
	if err != nil || revision != 4 {
		t.Fatalf("%s worker-first prepare=%d: %v",
			fixture.tenant.schema, revision, err)
	}
	live := beginProfileRuntimeTransaction(
		t, raceCtx, database, fixture.tenant.schema,
	)
	defer live.Rollback()
	livePID := maintenancePG18ProfileRacePID(t, raceCtx, live)
	result := make(chan profileRaceCallResult, 1)
	go func() {
		page, liveRevision, liveErr := executeMaintenancePG18RaceLiveWrite(
			raceCtx, live, fixture, tenantIndex, 2,
		)
		if liveErr == nil {
			liveErr = live.Commit()
		} else {
			_ = live.Rollback()
		}
		result <- profileRaceCallResult{
			page: page, revision: liveRevision, err: liveErr,
		}
	}()
	waitMaintenancePG18ProfileRaceLock(
		t, raceCtx, database, livePID, "live after worker",
	)
	if err = worker.Rollback(); err != nil {
		t.Fatalf("%s worker-first rollback: %v", fixture.tenant.schema, err)
	}
	liveResult := <-result
	if liveResult.err != nil || liveResult.page != 2 ||
		liveResult.revision != 4 {
		t.Fatalf("%s worker-first live=%d/%d: %v",
			fixture.tenant.schema, liveResult.page,
			liveResult.revision, liveResult.err)
	}
}

func maintenancePG18ProfileRacePID(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
) int {
	t.Helper()
	var pid int
	if err := transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.pg_backend_pid()").Scan(&pid); err != nil {
		t.Fatal(err)
	}
	return pid
}

func waitMaintenancePG18ProfileRaceLock(
	t *testing.T, ctx context.Context, database *sql.DB, pid int, label string,
) {
	t.Helper()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		err := database.QueryRowContext(ctx, `SELECT
coalesce(wait_event_type='Lock',false) FROM pg_catalog.pg_stat_activity
WHERE pid=$1`, pid).Scan(&waiting)
		if err != nil {
			t.Fatalf("%s race wait state: %v", label, err)
		}
		if waiting {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("%s did not block: %v", label, ctx.Err())
		case <-ticker.C:
		}
	}
}
