package portaluser

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.Create(context.Background(), platformtx.TenantScope{}, CreateCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Create error = %v, want ErrInvalid", err)
	}
}
