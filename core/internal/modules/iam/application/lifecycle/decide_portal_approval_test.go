package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestDecideRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.DecidePortalApproval(
		context.Background(), platformtx.TenantScope{}, PortalDecisionCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
