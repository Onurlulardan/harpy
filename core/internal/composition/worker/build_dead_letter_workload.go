package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/deadletterdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildDeadLetterWorkload(
	ctx context.Context,
	pool *postgres.Pool,
	settings config.RabbitMQ,
	catalogs platformasync.EventCatalogs,
	gate *admission.Gate,
) (managedAsyncWorkload, bool, error) {
	consumers := catalogs.Consumers()
	if len(consumers.Policies()) == 0 {
		return nil, false, nil
	}
	if gate == nil {
		return nil, false, fmt.Errorf("dead-letter admission gate is required")
	}
	ingest, err := asyncpostgres.BindDeadLetterIngest(pool, consumers)
	if err != nil {
		return nil, false, fmt.Errorf("bind dead-letter ingest store: %w", err)
	}
	connector, err := deadletterdelivery.NewRabbitConnector(settings, consumers, ingest)
	if err != nil {
		return nil, false, fmt.Errorf("build dead-letter connector: %w", err)
	}
	if err := probeDeadLetterDelivery(ctx, connector); err != nil {
		return nil, false, err
	}
	runner, err := deadletterdelivery.New(connector, deadletterdelivery.Settings{
		ReconnectMinimum: settings.ReconnectMinimum(),
		ReconnectMaximum: settings.ReconnectMaximum(),
	}, gate)
	if err != nil {
		return nil, false, fmt.Errorf("build dead-letter runner: %w", err)
	}
	return runner, true, nil
}
