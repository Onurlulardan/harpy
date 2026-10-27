package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/outboxdelivery"
)

func probeRabbitMQ(ctx context.Context, connector outboxdelivery.Connector) error {
	broker, err := connector.Open(ctx)
	if err != nil {
		return fmt.Errorf("RabbitMQ dependency preflight: %w", err)
	}
	if err := broker.Close(); err != nil {
		return fmt.Errorf("close RabbitMQ dependency preflight: %w", err)
	}
	return nil
}
