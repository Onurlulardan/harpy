package eventsv1

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestEnvelopeIsCanonicalDefensiveAndCataloged(t *testing.T) {
	catalog := testProducerCatalog(t)
	input := envelopeInput()
	value, err := NewEnvelope(input, catalog)
	if err != nil {
		t.Fatal(err)
	}
	input.Payload[2] = 'x'
	*input.TenantID = repairEventID
	if string(value.Payload) != `{"a":"value","z":1}` || *value.TenantID != eventTenantID {
		t.Fatal("event envelope was not canonical and defensively copied")
	}
	if value.PartitionKey() != eventTenantID+"/form_submission/"+aggregateID {
		t.Fatal("event partition key is not aggregate-stable")
	}
	document, _ := json.Marshal(value)
	if _, err := DecodeEnvelope(document, catalog); err != nil {
		t.Fatal(err)
	}
	badContract := bytes.Replace(document, []byte(EnvelopeContractVersion), []byte("integration-event.v2"), 1)
	if _, err := DecodeEnvelope(badContract, catalog); err == nil {
		t.Fatal("unknown envelope contract version was accepted")
	}
	unknown := bytes.Replace(document, []byte("{"), []byte(`{"unknown":true,`), 1)
	if _, err := DecodeEnvelope(unknown, catalog); err == nil {
		t.Fatal("unknown event envelope field was accepted")
	}
	unsupported := envelopeInput()
	unsupported.EventVersion = 2
	if _, err := NewEnvelope(unsupported, catalog); err == nil {
		t.Fatal("unregistered event version was accepted")
	}
}

func TestConsumerCatalogIsDeterministicAndQueueSafe(t *testing.T) {
	first, _ := NewConsumerPolicy("zeta.worker", 1, "workflow", "form.submitted", 1, "zeta.form.v1", EffectExternal, 4)
	second, _ := NewConsumerPolicy("alpha.worker", 1, "workflow", "form.submitted", 1, "alpha.form.v1", EffectExternal, 4)
	input := []ConsumerPolicy{first, second}
	catalog, err := NewConsumerCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := catalog.SHA256()
	input[0].Consumer = "mutated.worker"
	copy := catalog.Policies()
	copy[0].Consumer = "also.mutated"
	after, _ := catalog.SHA256()
	if before != after || catalog.Policies()[0].Consumer != "alpha.worker" {
		t.Fatal("consumer catalog was not deterministic and defensive")
	}
	second.QueueName = first.QueueName
	if _, err := NewConsumerCatalog([]ConsumerPolicy{first, second}); err == nil {
		t.Fatal("queue reuse across consumers was accepted")
	}
	anotherType, _ := NewConsumerPolicy(
		"zeta.worker", 1, "workflow", "form.deleted", 1, first.QueueName, EffectExternal, 4,
	)
	if _, err := NewConsumerCatalog([]ConsumerPolicy{first, anotherType}); err != nil {
		t.Fatalf("same consumer version could not share its queue: %v", err)
	}
	anotherType.DeliveryLimit = 5
	if _, err := NewConsumerCatalog([]ConsumerPolicy{first, anotherType}); err == nil {
		t.Fatal("one consumer queue accepted different delivery limits")
	}
	versionTwo, _ := NewConsumerPolicy(
		"zeta.worker", 2, "workflow", "form.submitted", 1, "zeta.form.v2", EffectExternal, 4,
	)
	versioned, err := NewConsumerCatalog([]ConsumerPolicy{first, versionTwo})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := versioned.Resolve("zeta.worker", 2, "form.submitted", 1); err != nil {
		t.Fatal("consumer version was not part of the subscription key")
	}
	versionTwo.QueueName = first.QueueName
	if _, err := NewConsumerCatalog([]ConsumerPolicy{first, versionTwo}); err == nil {
		t.Fatal("different consumer versions shared one durable queue")
	}
}

func TestBootstrapRequiresRegisteredPolicyAndStateWatermark(t *testing.T) {
	tenant := eventTenantID
	input := BootstrapReceipt{
		ReceiptID: receiptID, Consumer: "workflow.projector",
		ConsumerVersion: 1, TenantID: &tenant, ProducerAuthority: producerAuthority(),
		ConsumerAuthority: consumerAuthority(), AggregateType: "form_submission",
		AggregateID: aggregateID, ThroughSequence: 0, StateReference: "snapshot/submission/42",
		StateSHA256: digest("b"), CorrelationID: "bootstrap:42", CausationID: causationID,
		OccurredAt: eventTime,
	}
	if _, err := NewBootstrapReceipt(input, testConsumerCatalog(t, EffectExternal)); err != nil {
		t.Fatal(err)
	}
	input.ThroughSequence = -1
	if _, err := NewBootstrapReceipt(input, testConsumerCatalog(t, EffectExternal)); err == nil {
		t.Fatal("negative bootstrap watermark was accepted")
	}
	input.ThroughSequence, input.ConsumerVersion = 0, 2
	if _, err := NewBootstrapReceipt(input, testConsumerCatalog(t, EffectExternal)); err == nil {
		t.Fatal("bootstrap ignored the logical consumer version")
	}
}

func TestEffectReceiptMatchesEventAndEffectClass(t *testing.T) {
	externalCatalog := testConsumerCatalog(t, EffectExternal)
	producers := testProducerCatalog(t)
	envelope, _ := NewEnvelope(envelopeInput(), producers)
	external := effectReceiptInput()
	external.ExternalEffectReceiptID = eventString(effectID)
	if _, err := NewEffectReceipt(external, envelope, producers, externalCatalog); err != nil {
		t.Fatal(err)
	}
	missing := effectReceiptInput()
	if _, err := NewEffectReceipt(missing, envelope, producers, externalCatalog); err == nil {
		t.Fatal("external event effect without receipt was accepted")
	}
	databaseCatalog := testConsumerCatalog(t, EffectDatabaseTransactional)
	if _, err := NewEffectReceipt(effectReceiptInput(), envelope, producers, databaseCatalog); err != nil {
		t.Fatal(err)
	}
	wrongSequence := external
	wrongSequence.ThroughSequence = 8
	if _, err := NewEffectReceipt(wrongSequence, envelope, producers, externalCatalog); err == nil {
		t.Fatal("effect receipt for wrong sequence was accepted")
	}
}

func TestEventOperatorActionsPreserveRequiredIdentity(t *testing.T) {
	tenant := eventTenantID
	base := OperatorReceipt{
		ReceiptID: receiptID, Consumer: "workflow.projector", ConsumerVersion: 1, TenantID: &tenant,
		ProducerAuthority: producerAuthority(), ConsumerAuthority: consumerAuthority(), SourceEventID: eventID,
		ResultEventID: eventID, SourceBusinessEffectKey: "submission/42",
		ResultBusinessEffectKey: "submission/42", Reason: "verified",
		TicketReference: "OPS-42", ActorID: actorID, CausationID: causationID, OccurredAt: eventTime,
	}
	base.Action = OperatorRetry
	consumers := testConsumerCatalog(t, EffectExternal)
	retry, err := NewOperatorReceipt(base, consumers)
	if err != nil {
		t.Fatal(err)
	}
	document, _ := json.Marshal(retry)
	unknownAction := bytes.Replace(document, []byte(`"retry"`), []byte(`"other"`), 1)
	if _, err := DecodeOperatorReceipt(unknownAction, consumers); err == nil {
		t.Fatal("unknown event operator action was accepted")
	}
	base.Action = OperatorRepairReissue
	base.ResultEventID = repairEventID
	if _, err := NewOperatorReceipt(base, consumers); err != nil {
		t.Fatal(err)
	}
	base.ResultEventID = eventID
	if _, err := NewOperatorReceipt(base, consumers); err == nil {
		t.Fatal("repair reissue reused event ID")
	}
	base.ResultEventID = repairEventID
	base.ResultBusinessEffectKey = "changed-effect"
	if _, err := NewOperatorReceipt(base, consumers); err == nil {
		t.Fatal("repair reissue changed the business effect key")
	}
}

func TestEventUTF8ByteAndCorrelationBounds(t *testing.T) {
	input := envelopeInput()
	input.IdempotencyKey = strings.Repeat("ş", 256)
	if _, err := NewEnvelope(input, testProducerCatalog(t)); err != nil {
		t.Fatal(err)
	}
	input.IdempotencyKey += "ş"
	if _, err := NewEnvelope(input, testProducerCatalog(t)); err == nil {
		t.Fatal("event key over 512 UTF-8 bytes was accepted")
	}
	input = envelopeInput()
	input.CorrelationID = "request:allowed"
	if _, err := NewEnvelope(input, testProducerCatalog(t)); err != nil {
		t.Fatal(err)
	}
	input = envelopeInput()
	input.Payload = []byte(`{"a":1,"a":2}`)
	if _, err := NewEnvelope(input, testProducerCatalog(t)); err == nil {
		t.Fatal("duplicate event payload key was accepted")
	}
	input = envelopeInput()
	input.Payload = []byte(`{"value":"` + strings.Repeat("x", MaximumCanonicalPayloadBytes) + `"}`)
	if _, err := NewEnvelope(input, testProducerCatalog(t)); err == nil {
		t.Fatal("event payload over 65536 canonical bytes was accepted")
	}
}
