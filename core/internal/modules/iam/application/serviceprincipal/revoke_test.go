package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestRevokeRejectsUnboundTenantScope(t *testing.T) {
	command := RevokeCommand{
		GrantID: uuid.New(), ExpectedRevision: 1, Evidence: commandFixtureEvidence(),
	}
	_, err := commandFixtureCommands(t).Revoke(
		context.Background(), transaction.TenantScope{}, command,
	)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("unbound revoke error = %v", err)
	}
}
