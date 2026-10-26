package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestEnableRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.EnableWebAuthn(
		context.Background(), platformtx.TenantScope{}, WebAuthnStartCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
