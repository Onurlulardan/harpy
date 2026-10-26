package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/jackc/pgx/v5"
)

type observabilityTestDatabase struct {
	row       pgx.Row
	statement string
	arguments []any
}

func (database *observabilityTestDatabase) QueryRow(
	_ context.Context, statement string, arguments ...any,
) pgx.Row {
	database.statement = statement
	database.arguments = append([]any(nil), arguments...)
	return database.row
}

type observabilityTestRow struct {
	values [5]int64
	err    error
}

func (row observabilityTestRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errors.New("unexpected async observation destination count")
	}
	for index, destination := range destinations {
		value, ok := destination.(*int64)
		if !ok {
			return errors.New("unexpected async observation destination type")
		}
		*value = row.values[index]
	}
	return nil
}

func TestObserveAsyncBacklogUsesContractThreshold(t *testing.T) {
	database := &observabilityTestDatabase{
		row: observabilityTestRow{values: [5]int64{301, 1, 71, 2, 3}},
	}
	source := &observabilitySource{database: database}
	if _, err := source.Snapshot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if database.statement != observeAsyncBacklogSQL || len(database.arguments) != 1 ||
		database.arguments[0] != eventsv1.UndeliveredOutboxAlarmSeconds {
		t.Fatalf("async observation query contract = %q %#v",
			database.statement, database.arguments)
	}
}

func TestObserveAsyncBacklogRejectsInvalidDatabaseResult(t *testing.T) {
	source := &observabilitySource{database: &observabilityTestDatabase{
		row: observabilityTestRow{values: [5]int64{-1, 0, 0, 0, 0}},
	}}
	if _, err := source.Snapshot(context.Background()); err == nil {
		t.Fatal("invalid async observation result was accepted")
	}
}

func TestObserveAsyncBacklogSQLIsBoundedAndUsesDatabaseClock(t *testing.T) {
	for _, marker := range []string{
		"pg_catalog.clock_timestamp()", "platform.outbox", "platform.jobs",
		"event.delivery_state <> 'delivered'", "job.run_at <= observation_clock.observed_at",
		"$1::integer * interval '1 second'", "event.parked_at IS NOT NULL",
		"job.parked_at IS NOT NULL",
	} {
		if !strings.Contains(observeAsyncBacklogSQL, marker) {
			t.Fatalf("async observation SQL lacks %q", marker)
		}
	}
	if strings.Count(observeAsyncBacklogSQL, "pg_catalog.clock_timestamp()") != 1 ||
		strings.Count(observeAsyncBacklogSQL, ";") != 1 ||
		strings.Contains(observeAsyncBacklogSQL, "tenant_id") ||
		strings.Contains(strings.ToUpper(observeAsyncBacklogSQL), "GROUP BY") {
		t.Fatal("async observation SQL is not a one-row, low-cardinality statement")
	}
}
