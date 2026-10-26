//go:build integration

package rabbitmq

import (
	"context"
	"errors"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func TestRabbitMQ432ReliabilityContract(t *testing.T) {
	fixture := startRabbitFixture(t)
	connection := dialRabbit(t, fixture.settings)
	catalog := integrationConsumerCatalog(t)
	if err := DeclareTopology(connection, catalog); err != nil {
		t.Fatal(err)
	}
	publisher, err := OpenPublisher(connection)
	if err != nil {
		t.Fatal(err)
	}
	defer publisher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	shared := integrationEnvelope(t, "forms.entry.created", 1, testEventID)
	if _, err := publisher.Publish(ctx, PublishRequest{Envelope: shared, LeaseGeneration: 1}); err != nil {
		t.Fatal(err)
	}
	channel, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	first := getDelivery(t, channel, "workflow.projector.v1")
	second := getDelivery(t, channel, "analytics.projector.v1")
	if string(first.Body) != string(second.Body) || first.MessageId != testEventID {
		t.Fatal("separate subscriptions did not receive the same durable event")
	}
	_ = first.Ack(false)
	_ = second.Ack(false)

	unbound := integrationEnvelope(
		t, "forms.entry.unbound", 1, "66666666-6666-4666-8666-666666666666",
	)
	_, err = publisher.Publish(ctx, PublishRequest{Envelope: unbound, LeaseGeneration: 1})
	var failure *PublishFailure
	if !errors.As(err, &failure) || failure.Code() != FailureUnroutable {
		t.Fatalf("mandatory unroutable publish was not rejected: %v", err)
	}

	poison := integrationEnvelope(
		t, "forms.entry.created", 1, "77777777-7777-4777-8777-777777777777",
	)
	if _, err := publisher.Publish(ctx, PublishRequest{Envelope: poison, LeaseGeneration: 2}); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 3; attempt++ {
		delivery := getDelivery(t, channel, "workflow.projector.v1")
		if err := delivery.Reject(true); err != nil {
			t.Fatal(err)
		}
	}
	dead := getDelivery(t, channel, eventsv1.DeadLetterIngestQueue)
	if dead.MessageId != poison.EventID {
		t.Fatal("delivery-limit terminal did not reach durable dead-letter ingest")
	}
	_ = dead.Ack(false)

	restart := integrationEnvelope(
		t, "forms.entry.restart", 1, "88888888-8888-4888-8888-888888888888",
	)
	if _, err := publisher.Publish(ctx, PublishRequest{Envelope: restart, LeaseGeneration: 3}); err != nil {
		t.Fatal(err)
	}
	stopTimeout := 20 * time.Second
	if err := fixture.container.Stop(ctx, &stopTimeout); err != nil {
		t.Fatal(err)
	}
	if err := fixture.container.Start(ctx); err != nil {
		t.Fatal(err)
	}
	fixture.refresh(t)
	reconnected := dialRabbit(t, fixture.settings)
	restartChannel, err := reconnected.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer restartChannel.Close()
	persisted := getDelivery(t, restartChannel, "restart.projector.v1")
	if persisted.MessageId != restart.EventID {
		t.Fatal("quorum queue lost event across broker restart")
	}
	_ = persisted.Ack(false)
}

func integrationConsumerCatalog(t *testing.T) eventsv1.ConsumerCatalog {
	t.Helper()
	policies := []eventsv1.ConsumerPolicy{
		integrationConsumerPolicy(t, "workflow.projector", "workflow.projector.v1", "forms.entry.created", 2),
		integrationConsumerPolicy(t, "analytics.projector", "analytics.projector.v1", "forms.entry.created", 2),
		integrationConsumerPolicy(t, "restart.projector", "restart.projector.v1", "forms.entry.restart", 2),
	}
	catalog, err := eventsv1.NewConsumerCatalog(policies)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func integrationConsumerPolicy(
	t *testing.T, consumer string, queue string, eventType string, limit int,
) eventsv1.ConsumerPolicy {
	t.Helper()
	policy, err := eventsv1.NewConsumerPolicy(
		consumer, 1, "workflow", eventType, 1, queue, eventsv1.EffectExternal, limit,
	)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func integrationEnvelope(
	t *testing.T, eventType string, version int, eventID string,
) eventsv1.Envelope {
	t.Helper()
	base := validEnvelope(t)
	policy, err := eventsv1.NewProducerPolicy(eventType, version, "forms", base.PayloadSchemaSHA256)
	if err != nil {
		t.Fatal(err)
	}
	catalog, _ := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{policy})
	base.ContractVersion = ""
	base.EventID, base.EventType, base.EventVersion = eventID, eventType, version
	value, err := eventsv1.NewEnvelope(base, catalog)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
