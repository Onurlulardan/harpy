package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type cancelOwnedJobDatabase struct {
	state     string
	requested time.Time
	err       error
	statement string
	arguments []any
}

func (*cancelOwnedJobDatabase) Execute(
	context.Context,
	string,
	...any,
) (platformpostgres.CommandResult, error) {
	return platformpostgres.CommandResult{}, fmt.Errorf("unexpected execute")
}

func (database *cancelOwnedJobDatabase) QueryRow(
	_ context.Context,
	statement string,
	arguments ...any,
) platformpostgres.Row {
	database.statement = statement
	database.arguments = append([]any(nil), arguments...)
	return boundLifecycleRow{scan: func(destinations ...any) error {
		if database.err != nil {
			return database.err
		}
		*destinations[0].(*string) = database.state
		*destinations[1].(*time.Time) = database.requested
		return nil
	}}
}

func TestCancelOwnedJobUsesBoundAuthorityAndIsReplaySafe(t *testing.T) {
	tenantID := uuid.New()
	jobID := uuid.MustParse("019c6a7e-7c00-7000-8000-000000000030")
	requestedAt := time.Date(2026, 7, 17, 1, 2, 3, 456000000, time.UTC)
	request, err := NewOwnedJobCancellation(
		jobID, requestedAt, "BILLING_BOUNDARY_CANCELLED",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	)
	if err != nil {
		t.Fatal(err)
	}
	database := &cancelOwnedJobDatabase{state: "cancelled", requested: requestedAt}
	repository := &repository{database: database, authority: authorityStamp{
		scopeID: tenantID, groupID: "billing", groupGeneration: 3,
		domain: "billing", generation: 4,
	}}
	persisted, err := repository.CancelOwnedJob(context.Background(), request)
	if err != nil || !persisted.Equal(requestedAt) {
		t.Fatalf("cancel owned job = %v, %v", persisted, err)
	}
	if database.statement != cancelOwnedJobSQL || len(database.arguments) != 10 ||
		database.arguments[1] != tenantID || database.arguments[6] != jobID {
		t.Fatalf("cancel SQL binding = %q %#v", database.statement, database.arguments)
	}
	database.state = "running"
	if replay, replayErr := repository.CancelOwnedJob(
		context.Background(), request,
	); replayErr != nil || !replay.Equal(requestedAt) {
		t.Fatalf("cancel replay = %v, %v", replay, replayErr)
	}
}

func TestCancelOwnedJobFailsClosedBeforeOrAfterSQL(t *testing.T) {
	database := &cancelOwnedJobDatabase{err: pgx.ErrNoRows}
	repository := &repository{database: database, authority: authorityStamp{
		scopeID: uuid.New(), groupID: "billing", groupGeneration: 1,
		domain: "billing", generation: 1,
	}}
	if _, err := repository.CancelOwnedJob(
		context.Background(), OwnedJobCancellation{},
	); !errors.Is(err, ErrJobControlInvalid) || database.statement != "" {
		t.Fatalf("invalid cancellation = %v, SQL=%q", err, database.statement)
	}
	request, err := NewOwnedJobCancellation(
		uuid.MustParse("019c6a7e-7c00-7000-8000-000000000031"),
		time.Date(2026, 7, 17, 1, 2, 3, 0, time.UTC),
		"BILLING_BOUNDARY_CANCELLED",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.CancelOwnedJob(
		context.Background(), request,
	); !errors.Is(err, ErrJobNotControllable) {
		t.Fatalf("missing owned job error = %v", err)
	}
}
