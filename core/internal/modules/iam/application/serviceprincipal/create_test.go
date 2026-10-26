package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestCreateRejectsUnboundTenantScope(t *testing.T) {
	command := CreateCommand{
		ID: uuid.New(), Key: "workflow.invoice", DisplayName: "Invoice workflow",
		Evidence: commandFixtureEvidence(),
	}
	if command.Validate() != nil {
		t.Fatal("valid create command rejected")
	}
	_, err := commandFixtureCommands(t).Create(
		context.Background(), transaction.TenantScope{}, command,
	)
	if !errors.Is(err, ErrInvalidPrincipal) {
		t.Fatalf("unbound create error = %v", err)
	}
}
