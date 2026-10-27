package eventdelivery

import (
	"context"
	"strings"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

type fixtureHandler struct{}

func (*fixtureHandler) Handle(
	context.Context,
	eventsv1.Envelope,
	brokerrabbit.DeliveryMetadata,
) (brokerrabbit.HandleOutcome, error) {
	return brokerrabbit.HandleCommitted, nil
}

func fixturePolicy(
	t *testing.T,
	consumer string,
	version int,
	eventType string,
) eventsv1.ConsumerPolicy {
	t.Helper()
	policy, err := eventsv1.NewConsumerPolicy(
		consumer, version, "workflow", eventType, 1,
		consumer+".v"+string(rune('0'+version)),
		eventsv1.EffectDatabaseTransactional, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func fixtureCatalog(
	t *testing.T,
	policies ...eventsv1.ConsumerPolicy,
) eventsv1.ConsumerCatalog {
	t.Helper()
	catalog, err := eventsv1.NewConsumerCatalog(policies)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func fixtureProducers(
	t *testing.T,
	policies ...eventsv1.ConsumerPolicy,
) eventsv1.ProducerCatalog {
	t.Helper()
	seen := make(map[string]struct{})
	producers := make([]eventsv1.ProducerPolicy, 0, len(policies))
	for _, consumer := range policies {
		key := consumer.EventType + "\x00" + string(rune(consumer.EventVersion))
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		policy, err := eventsv1.NewProducerPolicy(
			consumer.EventType, consumer.EventVersion, "forms", strings.Repeat("a", 64),
		)
		if err != nil {
			t.Fatal(err)
		}
		producers = append(producers, policy)
	}
	catalog, err := eventsv1.NewProducerCatalog(producers)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func fixtureRegistration(
	t *testing.T,
	consumer string,
	version int,
) Registration {
	t.Helper()
	registration, err := NewRegistration(consumer, version, &fixtureHandler{})
	if err != nil {
		t.Fatal(err)
	}
	return registration
}

func openEventAdmission(t *testing.T) *admission.Gate {
	t.Helper()
	gate := admission.NewGate(nil)
	if !gate.Open() {
		t.Fatal("event admission did not open")
	}
	return gate
}
