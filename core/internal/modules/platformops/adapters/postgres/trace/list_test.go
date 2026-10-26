package trace

import (
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestListForwardsTraceFilter(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewTraceOperations(backend).ListTraces(
		t.Context(), "environment", platformopsapi.TraceFilter{})
	if err != nil || backend.operation != "list" || result.Total != 1 {
		t.Fatalf("ListTraces() = %#v, %v", result, err)
	}
}
