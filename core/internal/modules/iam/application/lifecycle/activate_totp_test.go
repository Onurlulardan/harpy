package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestActivateRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ActivateTOTP(
		context.Background(), platformtx.TenantScope{}, TOTPCompleteCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
