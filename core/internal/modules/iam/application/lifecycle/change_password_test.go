package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestChangeRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ChangePassword(
		context.Background(), platformtx.TenantScope{}, PasswordChangeCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
