package trace

import "testing"

func TestGetForwardsTraceLookup(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewTraceOperations(backend).GetTrace(t.Context(), "environment", "trace")
	if err != nil || backend.operation != "get" || len(result) != 1 {
		t.Fatalf("GetTrace() = %#v, %v", result, err)
	}
}
