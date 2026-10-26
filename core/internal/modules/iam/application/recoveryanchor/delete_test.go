package recoveryanchor

import (
	"context"
	"errors"
	"testing"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestDeleteRejectsMissingService(t *testing.T) {
	var service *Service
	var scope platformtx.TenantScope
	_, err := service.Delete(
		context.Background(),
		scope,
		anchordomain.DeleteCommand{},
	)
	if !errors.Is(err, anchordomain.ErrInvalid) {
		t.Fatalf("Delete() error = %v, want ErrInvalid", err)
	}
}
