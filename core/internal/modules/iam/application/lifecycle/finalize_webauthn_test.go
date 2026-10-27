package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestFinalizeRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.FinalizeWebAuthn(
		context.Background(), platformtx.TenantScope{}, WebAuthnCompleteCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
