package application

import "testing"

func TestTraceSpanLookupValidatesAndForwardsOwnedIdentity(t *testing.T) {
	fixture := &detailLookupFixture{}
	service := &OperationsService{traces: fixture}
	result, err := service.TraceSpanLookup(t.Context(), detailEnvironmentID, detailSpanID)
	if err != nil || fixture.operation != "trace_span" || result.SpanID != detailSpanID {
		t.Fatalf("TraceSpanLookup() = %#v, %v", result, err)
	}
	if _, err = service.TraceSpanLookup(t.Context(), detailEnvironmentID, "other"); err == nil {
		t.Fatal("TraceSpanLookup accepted an invalid span identity")
	}
}
