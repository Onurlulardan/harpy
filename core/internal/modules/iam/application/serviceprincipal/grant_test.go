package serviceprincipal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestGrantRejectsUnboundTenantScope(t *testing.T) {
	capability, err := ParseCapabilityKey("forms.record.update")
	if err != nil {
		t.Fatal(err)
	}
	command := GrantCommand{
		GrantID: uuid.New(), PrincipalID: uuid.New(), Capability: capability,
		Scope:          ResourceScope{Kind: ScopeTenant, Owner: "forms", ResourceKind: "record"},
		ValidFrom:      time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC),
		ManifestSHA256: [32]byte{3}, Evidence: commandFixtureEvidence(),
	}
	if command.Validate() != nil {
		t.Fatal("valid grant command rejected")
	}
	_, err = commandFixtureCommands(t).Grant(
		context.Background(), transaction.TenantScope{}, command,
	)
	if !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("unbound grant error = %v", err)
	}
}
