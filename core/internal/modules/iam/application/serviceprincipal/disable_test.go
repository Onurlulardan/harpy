package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestDisableRejectsUnboundTenantScope(t *testing.T) {
	_, err := commandFixtureCommands(t).Disable(
		context.Background(), transaction.TenantScope{}, commandFixtureState(),
	)
	if !errors.Is(err, ErrInvalidPrincipal) {
		t.Fatalf("unbound disable error = %v", err)
	}
}
