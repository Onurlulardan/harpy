package reports

import (
	"errors"
	"testing"

	queryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExecuteLegacyEventsRejectsUnboundService(t *testing.T) {
	service := &Service{}
	_, err := service.ExecuteLegacyEvents(
		t.Context(), platformtx.GlobalScope{}, queryapi.QueryRequest{},
	)
	if !errors.Is(err, ErrContract) {
		t.Fatalf("ExecuteLegacyEvents error = %v, want ErrContract", err)
	}
	if legacyEventSpec().Operation != AuditOperation {
		t.Fatal("legacy events must reuse the audited global read operation")
	}
}
