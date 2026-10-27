package eventsv1

import (
	"encoding/json"
	"testing"
)

func TestDecodeDeadLetterEnvelopeAcceptsValidUnregisteredVersion(t *testing.T) {
	catalog := testProducerCatalog(t)
	envelope, err := NewEnvelope(envelopeInput(), catalog)
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeDeadLetterEnvelope(document)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.EventID != envelope.EventID || decoded.PayloadSHA256 != envelope.PayloadSHA256 {
		t.Fatalf("decoded dead-letter envelope drifted: %#v", decoded)
	}
	if _, err := DecodeEnvelope(document, ProducerCatalog{}); err == nil {
		t.Fatal("normal decoder accepted an unregistered producer version")
	}
	if err := decoded.Validate(catalog); err != nil {
		t.Fatalf("decoded envelope lost registered validation: %v", err)
	}
}

func TestDecodeDeadLetterEnvelopeRejectsPayloadDigestMismatch(t *testing.T) {
	envelope := envelopeInput()
	envelope.ContractVersion = EnvelopeContractVersion
	envelope.PayloadSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	type wire Envelope
	document, err := json.Marshal(wire(envelope))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeDeadLetterEnvelope(document); err == nil {
		t.Fatal("dead-letter decoder accepted a payload digest mismatch")
	}
}
