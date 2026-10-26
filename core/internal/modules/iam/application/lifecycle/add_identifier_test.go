package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAddRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.AddIdentifier(
		context.Background(), platformtx.TenantScope{}, IdentifierAddCommand{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
