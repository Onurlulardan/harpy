//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"
)

const (
	terminalGapRetryInsertBarrierKey int64 = 725308704902
	terminalGapRedactionBarrierKey   int64 = 725308704903
)

type terminalGapBarrier struct {
	release func()
	cleanup func()
}

func installTerminalGapRetryInsertBarrier(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
) terminalGapBarrier {
	t.Helper()
	statement := fmt.Sprintf(`
CREATE FUNCTION platform.event_gap_terminal_retry_barrier_v1()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog
AS $function$
BEGIN
 PERFORM pg_catalog.pg_advisory_xact_lock(%d);
 RETURN NEW;
END;
$function$;
CREATE TRIGGER event_gap_terminal_retry_barrier
BEFORE INSERT ON platform.jobs
FOR EACH ROW WHEN (NEW.retry_of_job_id='%s'::uuid)
EXECUTE FUNCTION platform.event_gap_terminal_retry_barrier_v1();`,
		terminalGapRetryInsertBarrierKey, value.jobID,
	)
	return installTerminalGapBarrier(
		t, ctx, value.fixture.owner, terminalGapRetryInsertBarrierKey, statement,
		`DROP TRIGGER event_gap_terminal_retry_barrier ON platform.jobs;
DROP FUNCTION platform.event_gap_terminal_retry_barrier_v1();`,
	)
}

func installTerminalGapRedactionBarrier(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
) terminalGapBarrier {
	t.Helper()
	statement := fmt.Sprintf(`
CREATE FUNCTION platform.event_gap_terminal_redaction_barrier_v1()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog
AS $function$
BEGIN
 PERFORM pg_catalog.pg_advisory_xact_lock(%d);
 RETURN NEW;
END;
$function$;
CREATE TRIGGER event_gap_terminal_redaction_barrier
BEFORE UPDATE ON platform.dead_letters
FOR EACH ROW WHEN (
 OLD.id='%s'::uuid AND OLD.payload_canonical IS NOT NULL
 AND NEW.payload_canonical IS NULL
)
EXECUTE FUNCTION platform.event_gap_terminal_redaction_barrier_v1();`,
		terminalGapRedactionBarrierKey, value.rootID,
	)
	return installTerminalGapBarrier(
		t, ctx, value.fixture.owner, terminalGapRedactionBarrierKey, statement,
		`DROP TRIGGER event_gap_terminal_redaction_barrier ON platform.dead_letters;
DROP FUNCTION platform.event_gap_terminal_redaction_barrier_v1();`,
	)
}

func installTerminalGapBarrier(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	key int64,
	install string,
	drop string,
) terminalGapBarrier {
	t.Helper()
	if _, err := database.ExecContext(ctx, install); err != nil {
		t.Fatalf("install terminal gap barrier: %v", err)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(
		ctx, "SELECT pg_catalog.pg_advisory_lock($1)", key,
	); err != nil {
		connection.Close()
		t.Fatal(err)
	}
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() {
			_, _ = connection.ExecContext(
				context.Background(),
				"SELECT pg_catalog.pg_advisory_unlock($1)", key,
			)
			_ = connection.Close()
		})
	}
	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			release()
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if _, cleanupErr := database.ExecContext(cleanupCtx, drop); cleanupErr != nil {
				t.Errorf("drop terminal gap barrier: %v", cleanupErr)
			}
		})
	}
	return terminalGapBarrier{release: release, cleanup: cleanup}
}

func waitForTerminalGapRetryRootLock(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	waitForGapRaceCondition(t, ctx, fixture.owner, `
SELECT count(*) >= 2
FROM pg_catalog.pg_locks AS waiter
WHERE waiter.locktype='advisory' AND NOT waiter.granted
  AND waiter.database=(SELECT oid FROM pg_catalog.pg_database
                       WHERE datname=pg_catalog.current_database())`)
}
