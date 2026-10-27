package outboxdelivery

import (
	"context"
	"errors"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitConnector struct {
	settings  config.RabbitMQ
	consumers eventsv1.ConsumerCatalog
}

func NewRabbitConnector(
	settings config.RabbitMQ,
	consumers eventsv1.ConsumerCatalog,
) (*RabbitConnector, error) {
	validated, err := eventsv1.NewConsumerCatalog(consumers.Policies())
	if err != nil || settings.URL().Value() == "" {
		return nil, ErrDependencies
	}
	return &RabbitConnector{settings: settings, consumers: validated}, nil
}

func (connector *RabbitConnector) Open(ctx context.Context) (Broker, error) {
	if connector == nil || ctx == nil {
		return nil, ErrDependencies
	}
	connection, err := brokerrabbit.Dial(ctx, connector.settings)
	if err != nil {
		return nil, err
	}
	if err := brokerrabbit.DeclareTopology(connection, connector.consumers); err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("declare RabbitMQ topology: %w", err)
	}
	publisher, err := brokerrabbit.OpenPublisher(connection)
	if err != nil {
		_ = connection.Close()
		return nil, err
	}
	return &rabbitSession{connection: connection, publisher: publisher}, nil
}

type rabbitSession struct {
	connection *amqp091.Connection
	publisher  *brokerrabbit.Publisher
}

func (session *rabbitSession) Publish(
	ctx context.Context,
	request brokerrabbit.PublishRequest,
) (brokerrabbit.PublishReceipt, error) {
	return session.publisher.Publish(ctx, request)
}

func (session *rabbitSession) Close() error {
	return errors.Join(session.publisher.Close(), session.connection.Close())
}
