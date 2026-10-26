package analyticsdashboardexport

import (
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type Store struct {
	database platformpostgres.AsyncWorkerExecutor
}

func Bind(
	pool *platformpostgres.Pool,
) (platformasync.AnalyticsDashboardExportSource, error) {
	store, err := platformpostgres.BindAsyncWorkerExecutor(
		pool,
		func(executor platformpostgres.AsyncWorkerExecutor) *Store {
			return &Store{database: executor}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("bind analytics dashboard export source: %w", err)
	}
	return store, nil
}

var _ platformasync.AnalyticsDashboardExportSource = (*Store)(nil)
