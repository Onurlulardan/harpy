package postgres

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
)

func BindRetention(pool *asyncWorkerPool) (retention.Store, error) {
	store, err := bindAsyncWorkerStore(
		pool,
		func(database *asyncWorkerDatabase, _ cleanupDuration) retention.Store {
			return &retentionRepository{database: database}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("bind async retention store: %w", err)
	}
	return store, nil
}
