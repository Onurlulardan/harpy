package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRedeemRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.RedeemPortalSetup(
		context.Background(), platformtx.CredentialTenantScope{}, PortalSetupCompleteCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
