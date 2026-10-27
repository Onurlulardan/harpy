package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.GetSessions(
		context.Background(), platformtx.TenantScope{}, SessionListQuery{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
