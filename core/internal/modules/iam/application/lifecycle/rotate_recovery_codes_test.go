package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRotateRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.RotateRecoveryCodes(
		context.Background(), platformtx.TenantScope{}, RecoveryIssueCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
