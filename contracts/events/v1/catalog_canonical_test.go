package eventsv1

import (
	"bytes"
	"testing"
)

func TestCatalogJSONUsesRFC8785AndCarriesPayloadLimit(t *testing.T) {
	consumer := testConsumerCatalog(t, EffectExternal)
	consumerJSON, err := consumer.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(consumerJSON, []byte(`[{"broker_redelivery_horizon_days":`)) {
		t.Fatal("consumer catalog object keys are not RFC 8785 ordered")
	}

	producer := testProducerCatalog(t)
	policy := producer.Policies()[0]
	if policy.MaximumPayloadBytes != MaximumCanonicalPayloadBytes {
		t.Fatal("producer policy lost the exact canonical payload limit")
	}
	first, err := producer.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	second, err := producer.CanonicalJSON()
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("producer catalog canonical JSON is not deterministic")
	}
	policy.MaximumPayloadBytes++
	if err := policy.Validate(); err == nil {
		t.Fatal("producer policy accepted a different payload limit")
	}
}
