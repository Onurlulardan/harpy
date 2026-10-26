package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
	"github.com/jackc/pgx/v5"
)

type retentionTestDatabase struct {
	row       pgx.Row
	statement string
}

func (database *retentionTestDatabase) QueryRow(
	_ context.Context, statement string, _ ...any,
) pgx.Row {
	database.statement = statement
	return database.row
}

type retentionTestRow struct {
	acquired bool
	counts   [7]int32
	err      error
}

func (row retentionTestRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 8 {
		return errors.New("unexpected retention scan shape")
	}
	*destinations[0].(*bool) = row.acquired
	for index := range row.counts {
		*destinations[index+1].(*int32) = row.counts[index]
	}
	return nil
}

func TestRunRetentionCycleReturnsValidatedTypedCounts(t *testing.T) {
	database := &retentionTestDatabase{row: retentionTestRow{
		acquired: true, counts: [7]int32{1, 2, 3, 4, 5, 6, 7},
	}}
	repository := &retentionRepository{database: database}
	result, err := repository.RunRetentionCycle(context.Background())
	want := retention.StepCounts{
		JobsRedacted: 1, OutboxRedacted: 2, DeadLettersRedacted: 3,
		InboxPurged: 4, IdempotencyPurged: 5, JobsPurged: 6,
		DeadLettersPurged: 7,
	}
	if err != nil || !result.LockAcquired() || result.Counts() != want ||
		database.statement != runRetentionCycleSQL {
		t.Fatalf("retention result = %#v statement=%q: %v", result, database.statement, err)
	}
}

func TestRunRetentionCycleFailsClosed(t *testing.T) {
	if _, err := (*retentionRepository)(nil).RunRetentionCycle(context.Background()); !errors.Is(err, ErrRetentionStoreUnavailable) {
		t.Fatalf("nil repository error = %v", err)
	}
	if _, err := (&retentionRepository{}).RunRetentionCycle(nil); !errors.Is(err, ErrRetentionStoreUnavailable) {
		t.Fatalf("nil context error = %v", err)
	}
	database := &retentionTestDatabase{row: retentionTestRow{
		acquired: true, counts: [7]int32{501},
	}}
	if _, err := (&retentionRepository{database: database}).RunRetentionCycle(
		context.Background(),
	); !errors.Is(err, ErrRetentionResultInvalid) {
		t.Fatalf("invalid result error = %v", err)
	}
	want := errors.New("database unavailable")
	database.row = retentionTestRow{err: want}
	if _, err := (&retentionRepository{database: database}).RunRetentionCycle(
		context.Background(),
	); !errors.Is(err, want) {
		t.Fatalf("database error = %v", err)
	}
}
