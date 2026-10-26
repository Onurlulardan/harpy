package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExecutionCapabilitiesAreRedacted(t *testing.T) {
	port := testPort(func(transaction.TenantScope, Program) error { return nil })
	registration, _ := NewRegistration("forms", port)
	registry, _ := NewRegistry([]Registration{registration})
	program := newTestProgram(t, "forms", func(
		context.Context, SnapshotDriver,
	) error {
		return nil
	}, testScopeVerifier)
	values := []any{program, registration, registry}
	for _, value := range values {
		if fmt.Sprint(value) != executionRedaction {
			t.Fatalf("format leaked: %v", value)
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("JSON marshal accepted %T", value)
		}
	}
	if program.LogValue().String() != executionRedaction ||
		registration.LogValue().String() != executionRedaction ||
		registry.LogValue().String() != executionRedaction {
		t.Fatal("structured log value leaked")
	}
}
