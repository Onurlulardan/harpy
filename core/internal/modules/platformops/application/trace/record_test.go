package trace

import (
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestRecordForwardsTraceObservation(t *testing.T) {
	backend := &backendFixture{}
	err := NewTraceOperations(backend).RecordTrace(
		t.Context(), platformopsapi.TraceObservation{})
	if err != nil || backend.operation != "record" {
		t.Fatalf("RecordTrace() error = %v", err)
	}
}
