package transaction

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type Coordinator struct {
	pool *postgres.Pool
}

func NewCoordinator(pool *postgres.Pool) (*Coordinator, error) {
	if pool == nil {
		return nil, fmt.Errorf("tenant coordinator pool is nil")
	}
	profile := pool.StartupResult().Profile()
	if profile != postgres.ProfileAPI && profile != postgres.ProfileWorker {
		return nil, fmt.Errorf("tenant coordinator requires a runtime pool")
	}
	return &Coordinator{pool: pool}, nil
}
