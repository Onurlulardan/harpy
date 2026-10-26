package rabbitmq

import (
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	amqp "github.com/rabbitmq/amqp091-go"
)

type topologyChannel interface {
	ExchangeDeclare(string, string, bool, bool, bool, bool, amqp.Table) error
	QueueDeclare(string, bool, bool, bool, bool, amqp.Table) (amqp.Queue, error)
	QueueBind(string, string, string, bool, amqp.Table) error
	Close() error
}

func declarePublisherExchange(channel *amqp.Channel) error {
	return channel.ExchangeDeclare(
		eventsv1.BrokerExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

func DeclareTopology(connection *amqp.Connection, catalog eventsv1.ConsumerCatalog) error {
	if connection == nil || connection.IsClosed() {
		return fmt.Errorf("RabbitMQ topology connection is unavailable")
	}
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("open RabbitMQ topology channel: %w", err)
	}
	defer channel.Close()
	return declareTopology(channel, catalog)
}

func declareTopology(channel topologyChannel, catalog eventsv1.ConsumerCatalog) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ topology channel is unavailable")
	}
	if err := channel.ExchangeDeclare(
		eventsv1.BrokerExchangeName, "topic", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare event exchange: %w", err)
	}
	if err := channel.ExchangeDeclare(
		eventsv1.DeadLetterExchangeName, "topic", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare dead-letter exchange: %w", err)
	}
	if err := channel.ExchangeDeclare(
		eventsv1.DeadLetterQuarantineExchange,
		"direct", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare dead-letter quarantine exchange: %w", err)
	}
	if _, err := channel.QueueDeclare(
		eventsv1.DeadLetterIngestQueue, true, false, false, false,
		amqp.Table{
			"x-queue-type": "quorum", "x-delivery-limit": int32(-1),
			"x-dead-letter-exchange":    eventsv1.DeadLetterQuarantineExchange,
			"x-dead-letter-routing-key": eventsv1.DeadLetterQuarantineKey,
		},
	); err != nil {
		return fmt.Errorf("declare dead-letter ingest queue: %w", err)
	}
	if _, err := channel.QueueDeclare(
		eventsv1.DeadLetterQuarantineQueue, true, false, false, false,
		amqp.Table{
			"x-queue-type":  "quorum",
			"x-message-ttl": eventsv1.DeadLetterQuarantineTTLMS,
		},
	); err != nil {
		return fmt.Errorf("declare dead-letter quarantine queue: %w", err)
	}
	if err := channel.QueueBind(
		eventsv1.DeadLetterIngestQueue, "consumer.#",
		eventsv1.DeadLetterExchangeName, false, nil,
	); err != nil {
		return fmt.Errorf("bind dead-letter ingest queue: %w", err)
	}
	if err := channel.QueueBind(
		eventsv1.DeadLetterQuarantineQueue, eventsv1.DeadLetterQuarantineKey,
		eventsv1.DeadLetterQuarantineExchange, false, nil,
	); err != nil {
		return fmt.Errorf("bind dead-letter quarantine queue: %w", err)
	}
	return declareConsumerQueues(channel, catalog.Policies())
}

func declareConsumerQueues(channel topologyChannel, policies []eventsv1.ConsumerPolicy) error {
	declared := make(map[string]struct{}, len(policies))
	for _, policy := range policies {
		bindingKey, err := policy.BindingKey()
		if err != nil {
			return err
		}
		deadLetterKey, err := policy.DeadLetterRoutingKey()
		if err != nil {
			return err
		}
		if _, exists := declared[policy.QueueName]; !exists {
			arguments := amqp.Table{
				"x-queue-type": "quorum", "x-delivery-limit": int32(policy.DeliveryLimit),
				"x-dead-letter-exchange":    eventsv1.DeadLetterExchangeName,
				"x-dead-letter-routing-key": deadLetterKey,
			}
			if _, err := channel.QueueDeclare(
				policy.QueueName, true, false, false, false, arguments,
			); err != nil {
				return fmt.Errorf("declare consumer queue %s: %w", policy.QueueName, err)
			}
			declared[policy.QueueName] = struct{}{}
		}
		if err := channel.QueueBind(
			policy.QueueName, bindingKey, eventsv1.BrokerExchangeName, false, nil,
		); err != nil {
			return fmt.Errorf("bind consumer queue %s: %w", policy.QueueName, err)
		}
	}
	return nil
}
