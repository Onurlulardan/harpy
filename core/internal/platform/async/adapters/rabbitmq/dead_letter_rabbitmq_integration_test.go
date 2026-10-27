//go:build integration

package rabbitmq

import (
	"context"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRabbitMQ432DeadLetterIngestSettlement(t *testing.T) {
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
	envelope := integrationEnvelope(
		t, "forms.entry.created", 1,
		"99999999-9999-4999-8999-999999999999",
	)
	if _, err := publisher.Publish(ctx, PublishRequest{
		Envelope: envelope, LeaseGeneration: 1,
	}); err != nil {
		t.Fatal(err)
	}
	channel, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	for attempt := 0; attempt < 3; attempt++ {
		delivery := getDelivery(t, channel, "workflow.projector.v1")
		if err := delivery.Reject(true); err != nil {
			t.Fatal(err)
		}
	}
	dead := getDelivery(t, channel, eventsv1.DeadLetterIngestQueue)
	ingest := &rabbitDeadLetterCapture{}
	if err := settleDeadLetter(ctx, dead, catalog, ingest); err != nil {
		t.Fatal(err)
	}
	if ingest.root.RootID().String() == "00000000-0000-0000-0000-000000000000" ||
		ingest.root.Envelope().EventID != envelope.EventID ||
		ingest.root.ReasonClass() != deadletter.ReasonDeliveryLimit {
		t.Fatalf("real RabbitMQ dead-letter root = %#v", ingest.root)
	}
	inspection, err := channel.QueueInspect(eventsv1.DeadLetterIngestQueue)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Messages != 0 {
		t.Fatalf("persisted dead-letter remained unacked: %d", inspection.Messages)
	}
}

func TestRabbitMQ432MalformedDeadLetterIsQuarantined(t *testing.T) {
	fixture := startRabbitFixture(t)
	connection := dialRabbit(t, fixture.settings)
	catalog := integrationConsumerCatalog(t)
	if err := DeclareTopology(connection, catalog); err != nil {
		t.Fatal(err)
	}
	channel, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	body := []byte(`{"invalid":true}`)
	err = channel.PublishWithContext(
		ctx, eventsv1.DeadLetterExchangeName, "consumer.workflow.projector.v1",
		false, false, amqp.Publishing{
			ContentType:  eventsv1.BrokerContentType,
			DeliveryMode: amqp.Persistent, Body: body,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	delivery := getDelivery(t, channel, eventsv1.DeadLetterIngestQueue)
	if err := settleDeadLetter(ctx, delivery, catalog, &rabbitDeadLetterCapture{}); err != nil {
		t.Fatal(err)
	}
	quarantined := getDelivery(t, channel, eventsv1.DeadLetterQuarantineQueue)
	if string(quarantined.Body) != string(body) ||
		quarantined.RoutingKey != eventsv1.DeadLetterQuarantineKey {
		t.Fatalf("quarantined dead-letter drifted: key=%s body=%s",
			quarantined.RoutingKey, quarantined.Body)
	}
	if err := quarantined.Ack(false); err != nil {
		t.Fatal(err)
	}
}

type rabbitDeadLetterCapture struct {
	root deadletter.EventRoot
}

func (capture *rabbitDeadLetterCapture) IngestEvent(
	_ context.Context,
	root deadletter.EventRoot,
) (deadletter.IngestResult, error) {
	capture.root = root
	return deadletter.IngestResult{
		RootID: root.RootID(), Created: true, RecordedAt: time.Now().UTC(),
	}, nil
}
