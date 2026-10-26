package postgres

import (
	"context"
	"errors"
	"testing"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestBeginEventAssertsProducerBeforeReturningEffectPermit(t *testing.T) {
	producers, consumers := consumerCatalogs(t, "database_transactional")
	repository := consumerRepositoryFixture(producers, consumers)
	database := &consumerAuthorityDatabase{}
	repository.database = database
	envelope := consumerEnvelope(
		t, producers, "83000000-0000-4000-8000-000000000001",
		"83100000-0000-4000-8000-000000000001", 1, consumerTenantA,
	)
	decision, err := repository.BeginEvent(context.Background(), envelope)
	if decision != nil || !errors.Is(err, ErrConsumerAuthorityStale) {
		t.Fatalf("stale producer decision = %#v, %v", decision, err)
	}
	if len(database.operations) != 3 ||
		database.operations[0] != lockConsumerStreamSQL ||
		database.operations[1] != inspectEventSQL ||
		database.operations[2] != assertConsumerProducerSQL {
		t.Fatalf("event authority order = %#v", database.operations)
	}
}

type consumerAuthorityDatabase struct{ operations []string }

func (database *consumerAuthorityDatabase) Execute(
	_ context.Context,
	query string,
	_ ...any,
) (platformpostgres.CommandResult, error) {
	database.operations = append(database.operations, query)
	if query == assertConsumerProducerSQL {
		return platformpostgres.CommandResult{}, &pgconn.PgError{Code: "P5602"}
	}
	return platformpostgres.CommandResult{}, nil
}

func (database *consumerAuthorityDatabase) QueryRow(
	_ context.Context,
	query string,
	_ ...any,
) platformpostgres.Row {
	database.operations = append(database.operations, query)
	return consumerApplyRow{}
}

type consumerApplyRow struct{}

func (consumerApplyRow) Scan(destinations ...any) error {
	*destinations[0].(*string) = "apply"
	*destinations[1].(*int64) = 1
	return nil
}
