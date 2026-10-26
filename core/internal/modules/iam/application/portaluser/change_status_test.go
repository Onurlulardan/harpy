package portaluser

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestChangeRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ChangeStatus(context.Background(), platformtx.TenantScope{}, RevisionCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ChangeStatus error = %v, want ErrInvalid", err)
	}
}
