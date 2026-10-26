package adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ReceiveLegacyArchiveHit(ctx context.Context,
	scope platformtx.TenantScope, hit CompatibilityHit,
) (CompatibilityHit, error) {
	if application == nil || foundation.NilPort(application.binder) ||
		scope.OperationID() != "documents.compat.legacy_sync_archive" {
		return CompatibilityHit{}, ErrInvalid
	}
	store, err := application.binder.Bind(ctx, scope)
	if err != nil || store == nil {
		return CompatibilityHit{}, ErrUnavailable
	}
	bound := &Application{store: store, clock: application.clock}
	return bound.RecordCompatibilityHit(ctx, hit)
}
