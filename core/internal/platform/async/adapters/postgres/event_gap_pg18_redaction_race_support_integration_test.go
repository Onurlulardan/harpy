//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

const eventGapParkBarrierKey int64 = 725308704901

type eventGapRaceDelivery struct {
	outcome brokerrabbit.HandleOutcome
	err     error
}

type eventGapRaceRedaction struct {
	count int
	err   error
}

func ageOutboxForGapRedactionRace(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	eventID string,
) {
	t.Helper()
	transaction, err := fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(
		ctx, "ALTER TABLE platform.outbox DISABLE TRIGGER outbox_transition_guard",
	); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx, `
UPDATE platform.outbox
SET delivered_at=occurred_at+interval '1 hour',
 confirmed_at=occurred_at+interval '1 hour'
WHERE event_id=$1`, eventID); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(
		ctx, "ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_transition_guard",
	); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func installEventGapParkBarrier(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) func() {
	t.Helper()
	const statement = `
CREATE FUNCTION platform.event_gap_park_barrier_v1()
RETURNS trigger LANGUAGE plpgsql SET search_path=pg_catalog
AS $function$
BEGIN
 IF NEW.record_kind='root' AND NEW.reason_class='sequence_gap' THEN
  PERFORM pg_catalog.pg_advisory_xact_lock(725308704901);
 END IF;
 RETURN NEW;
END;
$function$;
CREATE TRIGGER event_gap_park_barrier
BEFORE INSERT ON platform.dead_letters
FOR EACH ROW EXECUTE FUNCTION platform.event_gap_park_barrier_v1();`
	if _, err := fixture.owner.ExecContext(ctx, statement); err != nil {
		t.Fatalf("install event gap park barrier: %v", err)
	}
	connection, err := fixture.owner.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(
		ctx, "SELECT pg_catalog.pg_advisory_lock($1)", eventGapParkBarrierKey,
	); err != nil {
		connection.Close()
		t.Fatal(err)
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			_, _ = connection.ExecContext(
				context.Background(), "SELECT pg_catalog.pg_advisory_unlock($1)",
				eventGapParkBarrierKey,
			)
			_ = connection.Close()
		})
	}
}

func waitForGapParkBarrier(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	waitForGapRaceCondition(t, ctx, fixture.owner, `
SELECT EXISTS (
 SELECT 1 FROM pg_catalog.pg_locks AS waiter
 WHERE waiter.locktype='advisory' AND NOT waiter.granted
   AND waiter.database=(SELECT oid FROM pg_catalog.pg_database
                        WHERE datname=pg_catalog.current_database())
)`)
}

func waitForGapRedactionLock(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	waitForGapRaceCondition(t, ctx, fixture.owner, `
SELECT EXISTS (
 SELECT 1 FROM pg_catalog.pg_stat_activity
 WHERE datname=pg_catalog.current_database() AND state='active'
   AND wait_event_type='Lock'
   AND query LIKE '%platform_async_run_retention_cycle_v1%'
)`)
}

func waitForGapRaceCondition(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	statement string,
) {
	t.Helper()
	waitCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var ready bool
		if err := database.QueryRowContext(waitCtx, statement).Scan(&ready); err != nil {
			t.Fatalf("observe event gap race: %v", err)
		}
		if ready {
			return
		}
		select {
		case <-waitCtx.Done():
			t.Fatal("event gap race did not reach the required lock boundary")
		case <-ticker.C:
		}
	}
}

func runOutboxRedactionForGapRace(
	ctx context.Context,
	fixture *eventGapPGFixture,
	result chan<- eventGapRaceRedaction,
) {
	counts, err := runRetentionCycleCounts(ctx, fixture.runtime)
	result <- eventGapRaceRedaction{count: int(counts[1]), err: err}
}
