package postgres

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
)

func BindDeadLetterIngest(
	pool *asyncWorkerPool,
	consumers eventsv1.ConsumerCatalog,
) (deadletter.IngestStore, error) {
	if len(consumers.Policies()) == 0 {
		return nil, ErrDeadLetterInputInvalid
	}
	store, err := bindAsyncWorkerStore(
		pool,
		func(database *asyncWorkerDatabase, cleanup cleanupDuration) deadletter.IngestStore {
			return &deadLetterRepository{
				database: database, cleanup: cleanup, consumers: consumers,
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("bind dead-letter ingest store: %w", err)
	}
	return store, nil
}

func BindDeadLetterOperator(
	pool *asyncWorkerPool,
	permissions deadletter.PermissionVerifier,
	identifiers deadletter.IDSource,
) (*deadletter.Operator, error) {
	store, err := bindAsyncWorkerStore(
		pool,
		func(database *asyncWorkerDatabase, cleanup cleanupDuration) deadletter.Store {
			return &deadLetterRepository{database: database, cleanup: cleanup}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("bind dead-letter operator store: %w", err)
	}
	operator, err := deadletter.NewOperator(permissions, store, identifiers)
	if err != nil {
		return nil, err
	}
	return operator, nil
}
