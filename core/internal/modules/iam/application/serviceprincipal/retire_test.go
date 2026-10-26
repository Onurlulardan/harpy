package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRetireRejectsUnboundTenantScope(t *testing.T) {
	_, err := commandFixtureCommands(t).Retire(
		context.Background(), transaction.TenantScope{}, commandFixtureState(),
	)
	if !errors.Is(err, ErrInvalidPrincipal) {
		t.Fatalf("unbound retire error = %v", err)
	}
}
