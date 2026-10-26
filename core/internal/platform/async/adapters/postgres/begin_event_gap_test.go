package postgres

import (
	"context"
	"errors"
	"testing"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func TestBeginEventReturnsParkedOnlyAfterDurableGapAndJob(t *testing.T) {
	producers, consumers := consumerCatalogs(t, "database_transactional")
	repository := consumerRepositoryFixture(producers, consumers)
	database := &consumerGapDatabase{}
	repository.database = database
	envelope := consumerEnvelope(
		t, producers, "84000000-0000-4000-8000-000000000001",
		"84100000-0000-4000-8000-000000000001", 2, consumerTenantA,
	)
	decision, err := repository.BeginEvent(context.Background(), envelope)
	gap, ok := decision.(ParkedSequenceGap)
	if err != nil || !ok || !gap.Created() || gap.EventID() != envelope.EventID ||
		gap.ExpectedSequence() != 1 || gap.ReceivedSequence() != 2 ||
		uuid.Validate(gap.RootID()) != nil || uuid.Validate(gap.ReconciliationJobID()) != nil {
		t.Fatalf("durable gap decision = %#v, %v", decision, err)
	}
	want := []string{
		lockConsumerStreamSQL, inspectEventSQL, assertConsumerProducerSQL, parkEventGapSQL,
	}
	if len(database.operations) != len(want) {
		t.Fatalf("gap operation count = %d", len(database.operations))
	}
	for index := range want {
		if database.operations[index] != want[index] {
			t.Fatalf("gap operation %d was not durable-first", index)
		}
	}
}

func TestBeginEventDoesNotParkWhenDurableWriteFails(t *testing.T) {
	producers, consumers := consumerCatalogs(t, "database_transactional")
	repository := consumerRepositoryFixture(producers, consumers)
	repository.database = &consumerGapDatabase{parkErr: errors.New("commit path unavailable")}
	envelope := consumerEnvelope(
		t, producers, "84200000-0000-4000-8000-000000000001",
		"84300000-0000-4000-8000-000000000001", 2, consumerTenantA,
	)
	decision, err := repository.BeginEvent(context.Background(), envelope)
	if decision != nil || err == nil {
		t.Fatalf("failed durable gap returned %#v, %v", decision, err)
	}
}

type consumerGapDatabase struct {
	operations []string
	parkErr    error
}

func (database *consumerGapDatabase) Execute(
	_ context.Context,
	query string,
	_ ...any,
) (platformpostgres.CommandResult, error) {
	database.operations = append(database.operations, query)
	return platformpostgres.CommandResult{}, nil
}

func (database *consumerGapDatabase) QueryRow(
	_ context.Context,
	query string,
	arguments ...any,
) platformpostgres.Row {
	database.operations = append(database.operations, query)
	if query == inspectEventSQL {
		return consumerGapRow{scan: func(values ...any) error {
			*values[0].(*string) = "gap"
			*values[1].(*int64) = 1
			return nil
		}}
	}
	return consumerGapRow{scan: func(values ...any) error {
		if database.parkErr != nil {
			return database.parkErr
		}
		*values[0].(*string) = arguments[0].(uuid.UUID).String()
		*values[1].(*string) = arguments[1].(uuid.UUID).String()
		*values[2].(*bool) = true
		return nil
	}}
}

type consumerGapRow struct{ scan func(...any) error }

func (row consumerGapRow) Scan(values ...any) error { return row.scan(values...) }
