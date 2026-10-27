package deadletterdelivery

import (
	"context"
	"errors"
	"fmt"
	"sync"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitConnector struct {
	settings  config.RabbitMQ
	consumers eventsv1.ConsumerCatalog
	ingest    deadletter.IngestStore
}

func NewRabbitConnector(
	settings config.RabbitMQ,
	consumers eventsv1.ConsumerCatalog,
	ingest deadletter.IngestStore,
) (*RabbitConnector, error) {
	validated, err := eventsv1.NewConsumerCatalog(consumers.Policies())
	if err != nil || len(validated.Policies()) == 0 ||
		settings.URL().Value() == "" || ingest == nil {
		return nil, ErrDependencies
	}
	return &RabbitConnector{settings: settings, consumers: validated, ingest: ingest}, nil
}

func (connector *RabbitConnector) Open(ctx context.Context) (Session, error) {
	if connector == nil || ctx == nil {
		return nil, ErrDependencies
	}
	connection, err := brokerrabbit.Dial(ctx, connector.settings)
	if err != nil {
		return nil, err
	}
	if err := brokerrabbit.DeclareTopology(connection, connector.consumers); err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("declare RabbitMQ dead-letter topology: %w", err)
	}
	consumer, err := brokerrabbit.OpenDeadLetterConsumer(
		connection, connector.consumers, connector.ingest,
	)
	if err != nil {
		_ = connection.Close()
		return nil, err
	}
	return &rabbitSession{connection: connection, consumer: consumer, close: &sync.Once{}}, nil
}

type rabbitSession struct {
	connection *amqp091.Connection
	consumer   *brokerrabbit.DeadLetterConsumer
	close      *sync.Once
	closeErr   error
}

func (session *rabbitSession) Run(ctx context.Context) error {
	if session == nil || session.consumer == nil {
		return ErrSessionUnavailable
	}
	return session.consumer.Run(ctx)
}

func (session *rabbitSession) Close() error {
	if session == nil || session.close == nil {
		return nil
	}
	session.close.Do(func() {
		session.closeErr = errors.Join(session.consumer.Close(), session.connection.Close())
	})
	return session.closeErr
}
