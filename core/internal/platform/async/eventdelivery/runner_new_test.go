package eventdelivery

import (
	"context"
	"errors"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestNewBuildsEmptyNoOpWithoutBrokerCredential(t *testing.T) {
	consumers, _ := eventsv1.NewConsumerCatalog(nil)
	producers, _ := eventsv1.NewProducerCatalog(nil)
	registry, _ := NewRegistry(consumers, nil)
	runner, err := New(config.RabbitMQ{}, producers, registry, openEventAdmission(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := runner.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if failure, open := <-runner.Failures(); open || failure != nil {
		t.Fatalf("empty runner emitted a failure: %v", failure)
	}
	if !errors.Is(runner.Start(context.Background()), ErrRunnerStarted) {
		t.Fatal("runner allowed a second start")
	}
}

func TestNewRequiresProducerCoverageBeforeBrokerUse(t *testing.T) {
	policy := fixturePolicy(t, "workflow.projector", 1, "forms.entry.created")
	consumers := fixtureCatalog(t, policy)
	registry, _ := NewRegistry(
		consumers, []Registration{fixtureRegistration(t, policy.Consumer, policy.ConsumerVersion)},
	)
	emptyProducers, _ := eventsv1.NewProducerCatalog(nil)
	if _, err := New(config.RabbitMQ{}, emptyProducers, registry, openEventAdmission(t)); err == nil {
		t.Fatal("consumer without producer contract was accepted")
	}
	if _, err := New(
		config.RabbitMQ{}, fixtureProducers(t, policy), registry, openEventAdmission(t),
	); !errors.Is(err, ErrRunnerDependencies) {
		t.Fatalf("non-empty registry accepted missing broker credential: %v", err)
	}
	if _, err := New(
		config.RabbitMQ{}, fixtureProducers(t, policy), registry, nil,
	); !errors.Is(err, ErrRunnerDependencies) {
		t.Fatalf("nil admission error=%v", err)
	}
}
