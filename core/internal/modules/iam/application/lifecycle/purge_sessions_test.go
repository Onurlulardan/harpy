package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestPurgeRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.PurgeSessions(
		context.Background(), platformtx.TenantScope{}, SessionRevokeAllCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
