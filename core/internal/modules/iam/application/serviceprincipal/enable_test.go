package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestEnableRejectsUnboundTenantScope(t *testing.T) {
	_, err := commandFixtureCommands(t).Enable(
		context.Background(), transaction.TenantScope{}, commandFixtureState(),
	)
	if !errors.Is(err, ErrInvalidPrincipal) {
		t.Fatalf("unbound enable error = %v", err)
	}
}
