package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCompleteRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.CompletePasswordReset(
		context.Background(), platformtx.CredentialTenantScope{}, PasswordResetCompleteCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
