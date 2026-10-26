package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReserveRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ReserveTOTP(
		context.Background(), platformtx.TenantScope{}, TOTPStartCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
