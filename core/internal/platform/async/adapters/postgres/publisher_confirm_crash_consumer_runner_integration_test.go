//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func startPipelineConsumer(
	t *testing.T,
	ctx context.Context,
	settings config.RabbitMQ,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
	handler *pipelineConsumerHandler,
) func() {
	t.Helper()
	connection := dialPipelineRabbit(t, ctx, settings)
	if err := brokerrabbit.DeclareTopology(connection, consumers); err != nil {
		_ = connection.Close()
		t.Fatalf("declare consumer topology: %v", err)
	}
	consumer, err := brokerrabbit.OpenConsumer(
		connection, producers, consumers, "projection.workflow", 1, handler,
	)
	if err != nil {
		_ = connection.Close()
		t.Fatalf("open pipeline consumer: %v", err)
	}
	runContext, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- consumer.Run(runContext) }()
	return func() {
		cancel()
		_ = consumer.Close()
		_ = connection.Close()
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, context.Canceled) {
				return
			}
		case <-time.After(5 * time.Second):
			t.Errorf("pipeline consumer did not stop")
		}
	}
}

func (handler *pipelineConsumerHandler) requireOutcome(
	t *testing.T,
	ctx context.Context,
	want brokerrabbit.HandleOutcome,
) {
	t.Helper()
	select {
	case err := <-handler.failures:
		t.Fatalf("pipeline consumer failed: %v", err)
	case got := <-handler.outcomes:
		if got != want {
			t.Fatalf("pipeline consumer outcome = %s, want %s", got, want)
		}
	case <-ctx.Done():
		t.Fatalf("wait for pipeline consumer %s: %v", want, ctx.Err())
	case <-time.After(90 * time.Second):
		t.Fatalf("pipeline consumer did not produce %s", want)
	}
}
