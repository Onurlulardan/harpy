package rabbitmq

import (
	"fmt"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestTopologyDeclaresVersionedQuorumQueuesAndDLX(t *testing.T) {
	first := consumerPolicy(t, "forms.entry.created", 1)
	second := consumerPolicy(t, "forms.entry.deleted", 2)
	catalog, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{second, first})
	if err != nil {
		t.Fatal(err)
	}
	channel := &fakeTopology{}
	if err := declareTopology(channel, catalog); err != nil {
		t.Fatal(err)
	}
	if len(channel.exchanges) != 3 || len(channel.queues) != 3 || len(channel.bindings) != 4 {
		t.Fatalf("unexpected topology: %+v", channel)
	}
	consumer := channel.queue("workflow.projector.v1")
	if consumer.arguments["x-queue-type"] != "quorum" ||
		consumer.arguments["x-delivery-limit"] != int32(8) ||
		consumer.arguments["x-dead-letter-exchange"] != eventsv1.DeadLetterExchangeName ||
		consumer.arguments["x-dead-letter-routing-key"] != "consumer.workflow.projector.v1" {
		t.Fatalf("consumer queue reliability arguments drifted: %+v", consumer.arguments)
	}
	ingest := channel.queue(eventsv1.DeadLetterIngestQueue)
	if ingest.arguments["x-delivery-limit"] != int32(-1) ||
		ingest.arguments["x-dead-letter-exchange"] != eventsv1.DeadLetterQuarantineExchange ||
		ingest.arguments["x-dead-letter-routing-key"] != eventsv1.DeadLetterQuarantineKey ||
		!channel.hasBinding(eventsv1.DeadLetterIngestQueue, "consumer.#", eventsv1.DeadLetterExchangeName) {
		t.Fatal("dead-letter ingest queue can drop terminal broker messages")
	}
	quarantine := channel.queue(eventsv1.DeadLetterQuarantineQueue)
	if quarantine.arguments["x-queue-type"] != "quorum" ||
		quarantine.arguments["x-message-ttl"] != eventsv1.DeadLetterQuarantineTTLMS ||
		!channel.hasBinding(
			eventsv1.DeadLetterQuarantineQueue, eventsv1.DeadLetterQuarantineKey,
			eventsv1.DeadLetterQuarantineExchange,
		) {
		t.Fatal("invalid dead-letter quarantine is not durable")
	}
	if !channel.hasBinding("workflow.projector.v1", "forms.entry.created.v1", eventsv1.BrokerExchangeName) ||
		!channel.hasBinding("workflow.projector.v1", "forms.entry.deleted.v2", eventsv1.BrokerExchangeName) {
		t.Fatal("consumer version bindings are incomplete")
	}
}

func consumerPolicy(t *testing.T, eventType string, eventVersion int) eventsv1.ConsumerPolicy {
	t.Helper()
	policy, err := eventsv1.NewConsumerPolicy(
		"workflow.projector", 1, "workflow", eventType, eventVersion,
		"workflow.projector.v1", eventsv1.EffectExternal, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

type exchangeDeclaration struct {
	name       string
	kind       string
	durable    bool
	autoDelete bool
}

type queueDeclaration struct {
	name      string
	durable   bool
	arguments amqp.Table
}

type bindingDeclaration struct {
	queue    string
	key      string
	exchange string
}

type fakeTopology struct {
	exchanges []exchangeDeclaration
	queues    []queueDeclaration
	bindings  []bindingDeclaration
}

func (channel *fakeTopology) ExchangeDeclare(
	name string, kind string, durable bool, autoDelete bool,
	_ bool, _ bool, _ amqp.Table,
) error {
	channel.exchanges = append(channel.exchanges, exchangeDeclaration{name, kind, durable, autoDelete})
	return nil
}

func (channel *fakeTopology) QueueDeclare(
	name string, durable bool, _ bool, _ bool, _ bool, arguments amqp.Table,
) (amqp.Queue, error) {
	copy := amqp.Table{}
	for key, value := range arguments {
		copy[key] = value
	}
	channel.queues = append(channel.queues, queueDeclaration{name, durable, copy})
	return amqp.Queue{Name: name}, nil
}

func (channel *fakeTopology) QueueBind(
	queue string, key string, exchange string, _ bool, _ amqp.Table,
) error {
	channel.bindings = append(channel.bindings, bindingDeclaration{queue, key, exchange})
	return nil
}

func (*fakeTopology) Close() error { return nil }

func (channel *fakeTopology) queue(name string) queueDeclaration {
	for _, declaration := range channel.queues {
		if declaration.name == name {
			return declaration
		}
	}
	panic(fmt.Sprintf("queue %s is not declared", name))
}

func (channel *fakeTopology) hasBinding(queue string, key string, exchange string) bool {
	for _, binding := range channel.bindings {
		if binding.queue == queue && binding.key == key && binding.exchange == exchange {
			return true
		}
	}
	return false
}
