package eventsv1

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSequenceGapEvidencePreservesExactCanonicalEnvelope(t *testing.T) {
	producers := testProducerCatalog(t)
	consumers := testConsumerCatalog(t, EffectDatabaseTransactional)
	input := envelopeInput()
	input.EventSequence = 3
	envelope, err := NewEnvelope(input, producers)
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := NewSequenceGapEvidence(
		envelope, "workflow.projector", 1, 2, producers, consumers,
	)
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeSequenceGapEvidence(document, producers, consumers)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, _ := json.Marshal(decoded)
	if !bytes.Equal(document, reencoded) || decoded.ReceivedSequence != 3 {
		t.Fatal("sequence gap evidence changed across strict decode")
	}
}

func TestSequenceGapEvidenceRejectsNonGapAndUnknownFields(t *testing.T) {
	producers := testProducerCatalog(t)
	consumers := testConsumerCatalog(t, EffectDatabaseTransactional)
	envelope, _ := NewEnvelope(envelopeInput(), producers)
	if _, err := NewSequenceGapEvidence(
		envelope, "workflow.projector", 1, envelope.EventSequence, producers, consumers,
	); err == nil {
		t.Fatal("non-gap event was accepted")
	}
	input := envelopeInput()
	input.EventSequence = 2
	envelope, _ = NewEnvelope(input, producers)
	evidence, _ := NewSequenceGapEvidence(
		envelope, "workflow.projector", 1, 1, producers, consumers,
	)
	document, _ := json.Marshal(evidence)
	unknown := bytes.Replace(document, []byte("{"), []byte(`{"unknown":true,`), 1)
	if _, err := DecodeSequenceGapEvidence(unknown, producers, consumers); err == nil {
		t.Fatal("unknown gap evidence field was accepted")
	}
}
