package effectsv1

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestReceiptSuccessAndFailureNullableMatrix(t *testing.T) {
	succeeded := successReceiptFixture(t)
	if err := succeeded.ValidateForRequest(requestFixture(t)); err != nil {
		t.Fatal(err)
	}

	providerReference := "provider-error/7"
	failed, err := NewFailedReceipt(
		requestFixture(t), testReceiptID, "RATE_LIMITED", ErrorTransient,
		&providerReference, "2026-07-15T10:00:01Z",
	)
	if err != nil {
		t.Fatal(err)
	}
	providerReference = "changed"
	if failed.ProviderReference == nil || *failed.ProviderReference != "provider-error/7" {
		t.Fatal("failed receipt retained caller provider-reference pointer")
	}

	invalidSuccess := succeeded
	class := ErrorPermanent
	invalidSuccess.ErrorClass = &class
	if err := invalidSuccess.Validate(); err == nil {
		t.Fatal("succeeded receipt accepted an error class")
	}
	invalidFailure := failed
	result := "forbidden-result"
	invalidFailure.ResultReference = &result
	if err := invalidFailure.Validate(); err == nil {
		t.Fatal("failed receipt accepted result evidence")
	}
}

func TestUnknownOutcomeCannotProduceReceipt(t *testing.T) {
	if _, err := NewFailedReceipt(
		requestFixture(t), testReceiptID, JobOutcomeUnknownCode, ErrorPermanent,
		nil, "2026-07-15T10:00:01Z",
	); err == nil {
		t.Fatal("unknown provider outcome produced a receipt")
	}
	if _, err := NewFailedReceipt(
		requestFixture(t), testReceiptID, "UNKNOWN", ErrorClass("unknown_outcome"),
		nil, "2026-07-15T10:00:01Z",
	); err == nil {
		t.Fatal("unknown outcome was accepted as an error class")
	}
	value := successReceiptFixture(t)
	value.Status = "unknown"
	if err := value.Validate(); err == nil {
		t.Fatal("unknown receipt status was accepted")
	}
}

func TestReceiptRoundTripRejectsUnknownWireValues(t *testing.T) {
	value := successReceiptFixture(t)
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeReceipt(document)
	if err != nil || decoded.ReceiptID != value.ReceiptID {
		t.Fatalf("receipt round trip failed: %v", err)
	}
	unknown := append([]byte(`{"unknown":true,`), document[1:]...)
	if _, err := DecodeReceipt(unknown); err == nil {
		t.Fatal("unknown receipt field was accepted")
	}
	if _, err := DecodeReceipt(append(document, []byte(` {}`)...)); err == nil {
		t.Fatal("trailing receipt document was accepted")
	}
	wrongVersion := bytes.Replace(document, []byte(ReceiptContractVersion), []byte("effect-receipt.v2"), 1)
	if _, err := DecodeReceipt(wrongVersion); err == nil {
		t.Fatal("unknown receipt version was accepted")
	}
	unknownStatus := bytes.Replace(document, []byte(ReceiptSucceeded), []byte("unknown"), 1)
	if _, err := DecodeReceipt(unknownStatus); err == nil {
		t.Fatal("unknown receipt status was accepted")
	}
}

func TestReceiptExactUTF8ByteLimitAndRequestIdentity(t *testing.T) {
	request := requestFixture(t)
	value := successReceiptFixture(t)
	value.Code = strings.Repeat("A", MaximumReceiptCodeBytes)
	if err := value.Validate(); err != nil {
		t.Fatalf("exact byte limit rejected: %v", err)
	}
	value.Code += "A"
	if err := value.Validate(); err == nil {
		t.Fatal("receipt code over byte limit was accepted")
	}
	value = successReceiptFixture(t)
	providerReference := strings.Repeat("é", MaximumProviderReferenceBytes/2)
	value.ProviderReference = &providerReference
	if err := value.Validate(); err != nil {
		t.Fatalf("exact UTF-8 provider-reference byte limit rejected: %v", err)
	}
	providerReference += "é"
	if err := value.Validate(); err == nil {
		t.Fatal("provider reference over UTF-8 byte limit was accepted")
	}
	value = successReceiptFixture(t)
	request.Operation = "different-operation"
	if err := value.ValidateForRequest(request); err == nil {
		t.Fatal("receipt accepted a different request identity")
	}
}
