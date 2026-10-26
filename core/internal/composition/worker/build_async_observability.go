package worker

import (
	"fmt"
	"log/slog"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/telemetry"
)

func buildAsyncObservability(
	pool *platformpostgres.Pool,
	runtime *telemetry.Runtime,
	logger *slog.Logger,
) error {
	source, err := asyncpostgres.BindObservability(pool)
	if err != nil {
		return fmt.Errorf("bind async observability: %w", err)
	}
	if err := runtime.RegisterAsyncBacklog(source, logger); err != nil {
		return fmt.Errorf("register async observability: %w", err)
	}
	return nil
}
