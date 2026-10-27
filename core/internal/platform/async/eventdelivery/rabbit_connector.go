package eventdelivery

import (
	"context"
	"errors"
	"fmt"
	"sync"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitConnector struct {
	settings  config.RabbitMQ
	producers eventsv1.ProducerCatalog
	consumers eventsv1.ConsumerCatalog
}

func (connector *rabbitConnector) Open(
	ctx context.Context,
	registration Registration,
) (brokerSession, error) {
	connection, err := brokerrabbit.Dial(ctx, connector.settings)
	if err != nil {
		return nil, err
	}
	if err := brokerrabbit.DeclareTopology(connection, connector.consumers); err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("declare RabbitMQ event topology: %w", err)
	}
	consumer, err := brokerrabbit.OpenConsumer(
		connection, connector.producers, connector.consumers,
		registration.consumer, registration.version, registration.handler,
	)
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("open RabbitMQ event consumer: %w", err)
	}
	return &rabbitSession{connection: connection, consumer: consumer}, nil
}

type rabbitSession struct {
	connection *amqp.Connection
	consumer   *brokerrabbit.Consumer
	closeOnce  sync.Once
	closeErr   error
}

func (session *rabbitSession) Run(ctx context.Context) error {
	consumerResult := make(chan error, 1)
	connectionClosed := session.connection.NotifyClose(make(chan *amqp.Error, 1))
	go func() { consumerResult <- session.consumer.Run(ctx) }()
	select {
	case err := <-consumerResult:
		return err
	case closeError, open := <-connectionClosed:
		_ = session.consumer.Close()
		if open && closeError != nil {
			return fmt.Errorf("RabbitMQ event connection closed: %w", closeError)
		}
		return fmt.Errorf("RabbitMQ event connection closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (session *rabbitSession) Close() error {
	session.closeOnce.Do(func() {
		session.closeErr = errors.Join(session.consumer.Close(), session.connection.Close())
	})
	return session.closeErr
}
