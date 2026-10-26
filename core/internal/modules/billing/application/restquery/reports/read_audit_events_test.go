package reports

import (
	"errors"
	"testing"

	queryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReadAuditEventsRejectsUnboundService(t *testing.T) {
	service := &Service{}
	_, err := service.ReadAuditEvents(
		t.Context(), platformtx.GlobalScope{}, queryapi.QueryRequest{},
	)
	if !errors.Is(err, ErrContract) {
		t.Fatalf("ReadAuditEvents error = %v, want ErrContract", err)
	}
}
