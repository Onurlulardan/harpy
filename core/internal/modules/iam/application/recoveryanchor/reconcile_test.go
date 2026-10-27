package recoveryanchor

import (
	"context"
	"errors"
	"testing"
	"time"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReconcileRejectsMissingService(t *testing.T) {
	var service *Service
	var scope platformtx.TenantScope
	_, err := service.Reconcile(
		context.Background(),
		scope,
		anchordomain.Subject{},
		anchordomain.ID{},
		time.Time{},
	)
	if !errors.Is(err, anchordomain.ErrInvalid) {
		t.Fatalf("Reconcile() error = %v, want ErrInvalid", err)
	}
}
