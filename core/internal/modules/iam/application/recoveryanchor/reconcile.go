package recoveryanchor

import (
	"context"
	"time"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) Reconcile(
	ctx context.Context,
	scope platformtx.TenantScope,
	subject anchordomain.Subject,
	operationID anchordomain.ID,
	now time.Time,
) (anchordomain.CommitResult, error) {
	if service == nil || !scopeValid(scope, subject) || operationID == [16]byte{} || now.IsZero() {
		return anchordomain.CommitResult{}, anchordomain.ErrInvalid
	}
	return service.store.Commit(ctx, scope, subject, operationID, now)
}
