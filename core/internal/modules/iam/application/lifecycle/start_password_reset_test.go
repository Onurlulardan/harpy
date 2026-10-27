package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestStartRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.StartPasswordReset(
		context.Background(), platformtx.CredentialTenantScope{}, PasswordResetStartCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
