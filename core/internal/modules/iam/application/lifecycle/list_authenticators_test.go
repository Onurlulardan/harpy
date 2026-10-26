package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ListAuthenticators(
		context.Background(), platformtx.TenantScope{}, AuthenticatorListQuery{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
