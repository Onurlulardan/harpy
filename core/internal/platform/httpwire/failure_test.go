package httpwire

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestWriteFailureUsesCanonicalEnvelope(t *testing.T) {
	recorder := httptest.NewRecorder()
	if err := WriteFailure(recorder, "request-1", FailureRouteNotFound, false); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	var envelope runtimev1.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != "ROUTE_NOT_FOUND" || envelope.Error.CorrelationID != "request-1" || envelope.Error.Issues == nil {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
}

func TestWriteFailureRejectsUnknownFailure(t *testing.T) {
	if err := WriteFailure(httptest.NewRecorder(), "request-1", Failure(255), false); err == nil {
		t.Fatal("expected unknown failure rejection")
	}
}
