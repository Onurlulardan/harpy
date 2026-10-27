package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAcceptRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.AcceptInitialAdminSetup(
		context.Background(),
		platformtx.CredentialTenantScope{},
		InitialAdminSetupAcceptCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
