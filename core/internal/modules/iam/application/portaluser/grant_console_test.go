package portaluser

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGrantRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.GrantConsole(context.Background(), platformtx.TenantScope{}, RevisionCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("GrantConsole error = %v, want ErrInvalid", err)
	}
}
