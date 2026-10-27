package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestResolvePortalSetupRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ResolvePortalSetup(
		context.Background(), platformtx.TenantScope{}, PortalSetupQuery{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
