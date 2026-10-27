package eventsv1

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestEnvelopeDecodeRejectsAmbiguousAndOversizedDocuments(t *testing.T) {
	catalog := testProducerCatalog(t)
	envelope, err := NewEnvelope(envelopeInput(), catalog)
	if err != nil {
		t.Fatal(err)
	}
	document, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := append([]byte(`{"contract_version":"integration-event.v1",`), document[1:]...)
	if _, err := DecodeEnvelope(duplicate, catalog); err == nil {
		t.Fatal("duplicate envelope field was accepted")
	}
	oversized := append(append([]byte(nil), document...), bytes.Repeat([]byte(" "), MaximumDocumentBytes)...)
	if _, err := DecodeEnvelope(oversized, catalog); err == nil {
		t.Fatal("oversized event document was accepted")
	}
	invalidUTF8 := []byte{'{', '"', 0xff, '"', ':', '1', '}'}
	if _, err := DecodeEnvelope(invalidUTF8, catalog); err == nil {
		t.Fatal("invalid UTF-8 event document was accepted")
	}
}
