package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
)

func TestHTTPFailureEnvelopeIsDeterministic(t *testing.T) {
	fixture := newLayoutHTTPFixture(t)
	fixture.access.err = layouts.ErrLayoutAccessForbidden
	request := layoutHTTPRequest(t, http.MethodGet, "")
	response := httptest.NewRecorder()
	fixture.handler.getRevision(response, request)
	if response.Code != http.StatusForbidden || fixture.app.getCalls != 0 ||
		response.Header().Get("X-Correlation-ID") != "layout-request-1" {
		t.Fatalf("failure status=%d headers=%v calls=%d",
			response.Code, response.Header(), fixture.app.getCalls)
	}
	var envelope layoutHTTPErrorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Success || envelope.Error.Code != "AUTHORIZATION_DENIED" ||
		envelope.Error.Message != "Operation is forbidden" ||
		envelope.Error.CorrelationID != "layout-request-1" ||
		envelope.Error.Issues == nil || len(envelope.Error.Issues) != 0 {
		t.Fatalf("failure envelope = %+v", envelope)
	}
}
