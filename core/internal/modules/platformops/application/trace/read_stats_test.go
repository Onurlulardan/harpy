package trace

import "testing"

func TestReadForwardsTraceStatistics(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewTraceOperations(backend).ReadTraceStats(t.Context(), "environment")
	if err != nil || backend.operation != "read" || result.TotalTraces != 1 {
		t.Fatalf("ReadTraceStats() = %#v, %v", result, err)
	}
}
