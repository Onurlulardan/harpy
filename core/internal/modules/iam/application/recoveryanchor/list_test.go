package recoveryanchor

import (
	"context"
	"errors"
	"testing"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListRejectsMissingService(t *testing.T) {
	var service *Service
	var scope platformtx.TenantScope
	_, err := service.List(
		context.Background(),
		scope,
		anchordomain.Subject{},
	)
	if !errors.Is(err, anchordomain.ErrInvalid) {
		t.Fatalf("List() error = %v, want ErrInvalid", err)
	}
}
