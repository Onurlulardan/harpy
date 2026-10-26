package portaluser

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRevokeRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.RevokeConsole(context.Background(), platformtx.TenantScope{}, RevisionCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("RevokeConsole error = %v, want ErrInvalid", err)
	}
}
