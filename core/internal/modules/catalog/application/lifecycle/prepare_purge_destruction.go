package lifecycle

import (
	"context"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) preparePurgeDestruction(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
	authority PurgeAuthority,
	now time.Time,
) (catalogdomain.Operation, time.Time, bool, error) {
	if operation.DestructiveStarted == nil {
		retryAt, ready, err := application.preflightPurgeOwners(
			ctx, scope, operation, authority,
		)
		if err != nil || !ready {
			return operation, retryAt, ready, err
		}
		next, startErr := operation.StartDestructive(now)
		if startErr != nil {
			return operation, time.Time{}, false, ErrConflict
		}
		operation, err = ChangeOperation(ctx, store, OperationMutation{
			Operation: next, ExpectedRevision: operation.Revision,
		})
		if err != nil {
			return operation, time.Time{}, false, err
		}
	}
	retryAt, ready, err := application.beginPurgeOwners(
		ctx, scope, operation, authority,
	)
	return operation, retryAt, ready, err
}
