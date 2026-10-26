//go:build integration

package eventdelivery

import (
	"context"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

func TestRabbitMQ432RunnerReconnectsAndSettlesDuplicate(t *testing.T) {
	fixture := startEventRabbit(t)
	producers, consumers := integrationEventContracts(t)
	handler := newIntegrationHandler()
	registration, err := NewRegistration("workflow.eventdelivery", 1, handler)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(consumers, []Registration{registration})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := New(fixture.settings, producers, registry, openEventAdmission(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { runner.BeginDrain(); _ = runner.Force() })

	connection := dialEventRabbit(t, fixture.settings)
	if err := brokerrabbit.DeclareTopology(connection, consumers); err != nil {
		t.Fatal(err)
	}
	publisher, err := brokerrabbit.OpenPublisher(connection)
	if err != nil {
		t.Fatal(err)
	}
	defer publisher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	duplicate := integrationEvent(
		t, producers, "11111111-1111-4111-8111-111111111111", 1,
	)
	publishIntegrationEvent(t, ctx, publisher, duplicate, 1)
	publishIntegrationEvent(t, ctx, publisher, duplicate, 2)
	if !handler.waitFor(2) {
		t.Fatal("event handler did not settle duplicate deliveries")
	}
	time.Sleep(100 * time.Millisecond)

	originalURL := fixture.settings.URL().Value()
	stopTimeout := 20 * time.Second
	if err := fixture.container.Stop(ctx, &stopTimeout); err != nil {
		t.Fatal(err)
	}
	if err := fixture.container.Start(ctx); err != nil {
		t.Fatal(err)
	}
	fixture.refresh(t)
	if fixture.settings.URL().Value() != originalURL {
		t.Fatal("RabbitMQ mapped endpoint changed across restart")
	}
	restarted := dialEventRabbit(t, fixture.settings)
	if err := brokerrabbit.DeclareTopology(restarted, consumers); err != nil {
		t.Fatal(err)
	}
	restartPublisher, err := brokerrabbit.OpenPublisher(restarted)
	if err != nil {
		t.Fatal(err)
	}
	defer restartPublisher.Close()
	afterRestart := integrationEvent(
		t, producers, "66666666-6666-4666-8666-666666666666", 2,
	)
	publishIntegrationEvent(t, ctx, restartPublisher, afterRestart, 3)
	if !handler.waitFor(3) {
		runner.lifecycle.mutex.Lock()
		active := len(runner.lifecycle.active)
		runner.lifecycle.mutex.Unlock()
		select {
		case failure := <-runner.Failures():
			t.Fatalf("event runner stopped after restart: %v", failure)
		default:
			t.Fatalf("event runner did not reconnect after restart; active sessions=%d", active)
		}
	}

	runner.BeginDrain()
	if err := runner.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if _, open := <-runner.Failures(); open {
		t.Fatal("clean RabbitMQ drain left failure stream open")
	}
	if got := handler.snapshot(); len(got) != 3 ||
		got[0] != brokerrabbit.HandleCommitted ||
		got[1] != brokerrabbit.HandleDuplicate ||
		got[2] != brokerrabbit.HandleCommitted {
		t.Fatalf("unexpected duplicate settlement sequence: %v", got)
	}
}

func publishIntegrationEvent(
	t *testing.T,
	ctx context.Context,
	publisher *brokerrabbit.Publisher,
	envelope eventsv1.Envelope,
	lease int64,
) {
	t.Helper()
	if _, err := publisher.Publish(ctx, brokerrabbit.PublishRequest{
		Envelope: envelope, LeaseGeneration: lease,
	}); err != nil {
		t.Fatal(err)
	}
}
