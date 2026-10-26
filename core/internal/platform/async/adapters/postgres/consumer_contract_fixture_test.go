package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

const (
	consumerTenantA = "11111111-1111-4111-8111-111111111111"
	consumerTenantB = "22222222-2222-4222-8222-222222222222"
	consumerActor   = "44444444-4444-4444-8444-444444444444"
	consumerCause   = "55555555-5555-4555-8555-555555555555"
)

func consumerCatalogs(
	t *testing.T,
	effect eventsv1.EffectClass,
) (eventsv1.ProducerCatalog, eventsv1.ConsumerCatalog) {
	t.Helper()
	producer, err := eventsv1.NewProducerPolicy(
		"workflow.changed", 1, "workflow", strings.Repeat("b", 64),
	)
	if err != nil {
		t.Fatal(err)
	}
	producers, err := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{producer})
	if err != nil {
		t.Fatal(err)
	}
	consumer, err := eventsv1.NewConsumerPolicy(
		"projection.workflow", 1, "workflow", "workflow.changed", 1,
		"projection.workflow.changed.v1", effect, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	consumers, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{consumer})
	if err != nil {
		t.Fatal(err)
	}
	return producers, consumers
}

func consumerEnvelope(
	t *testing.T,
	producers eventsv1.ProducerCatalog,
	eventID string,
	aggregateID string,
	sequence int64,
	tenant string,
) eventsv1.Envelope {
	t.Helper()
	payload := []byte(`{"ok":true}`)
	value, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID: eventID, EventType: "workflow.changed", EventVersion: 1,
		TenantID: &tenant, ProducerAuthority: consumerEventAuthority(tenant),
		AggregateType: "workflow.instance", AggregateID: aggregateID,
		AggregateVersion: sequence, EventSequence: sequence,
		OccurredAt: "2026-07-15T10:00:00Z", ActorID: consumerActor,
		CorrelationID: "consumer:test.1", CausationID: consumerCause,
		IdempotencyKey:               "event/" + eventID,
		BusinessEffectIdempotencyKey: "effect/" + eventID,
		PayloadSchemaSHA256:          strings.Repeat("b", 64),
		PayloadSHA256:                consumerDigest(payload), Payload: payload,
	}, producers)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func consumerEventAuthority(tenant string) eventsv1.Authority {
	return eventsv1.Authority{
		ScopeKind: eventsv1.ScopeTenant, ScopeID: tenant,
		GroupID: "operations", GroupGeneration: 1,
		Domain: "workflow", CutoverGeneration: 1, WriterOrigin: eventsv1.WriterV2,
	}
}

func consumerRepositoryFixture(
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
) *consumerRepository {
	gapPolicy, _ := platformasync.EventGapJobPolicy("workflow")
	return &consumerRepository{
		authority: authorityStamp{
			scopeID: uuid.MustParse(consumerTenantA), groupID: "operations",
			groupGeneration: 1, domain: "workflow", generation: 1,
		},
		gapPolicy: gapPolicy, producers: producers, consumers: consumers,
		consumer: "projection.workflow", consumerVersion: 1,
		identity: &consumerIdentity{marker: 1},
	}
}

func consumerJobCatalog(t *testing.T) jobsv1.Catalog {
	t.Helper()
	policy, err := platformasync.EventGapJobPolicy("workflow")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func consumerDigest(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func consumerEffectAuthority(tenant string) effectsv1.Authority {
	return effectsv1.Authority{
		ScopeKind: effectsv1.ScopeTenant, ScopeID: tenant,
		GroupID: "operations", GroupGeneration: 1,
		Domain: "workflow", CutoverGeneration: 1, WriterOrigin: effectsv1.WriterV2,
	}
}
