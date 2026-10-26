package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReadRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ReadRecoveryCodes(
		context.Background(), platformtx.TenantScope{}, RecoveryListQuery{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
