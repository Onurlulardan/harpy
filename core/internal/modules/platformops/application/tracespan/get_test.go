package tracespan

import (
	"context"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) TraceSpanLookup(
	context.Context, string, string,
) (platformopsapi.TraceSpan, error) {
	backend.called = true
	return platformopsapi.TraceSpan{TraceObservation: platformopsapi.TraceObservation{
		SpanID: "span"}}, nil
}

func TestGetForwardsTraceSpanLookup(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewTraceSpanOperations(backend).GetTraceSpan(
		t.Context(), "environment", "span")
	if err != nil || !backend.called || result.SpanID != "span" {
		t.Fatalf("GetTraceSpan() = %#v, %v", result, err)
	}
}
