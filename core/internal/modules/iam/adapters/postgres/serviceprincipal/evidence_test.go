package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestEvidenceReaderRejectsUnboundScope(t *testing.T) {
	port, err := NewEvidenceReader()
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = port.Evidence(
		context.Background(), platformtx.TenantScope{}, uuid.New(), 1,
	)
	if !errors.Is(err, errEvidenceRead) {
		t.Fatalf("unbound scope error = %v", err)
	}
}
