package systemtransfer

import (
	"context"
	"errors"
)

func FinalizeSourceCleanup(ctx context.Context,
	store SourceCleanupStore,
	target SourceCleanupTarget,
	deleted SourceDeleteResult,
) (bool, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(store) ||
		target.Validate() != nil {
		return false, ErrSourceCleanupInvalid
	}
	if deleted.Validate() != nil {
		return false, errors.Join(
			ErrSourceDeleteOutcomeUnknown, ErrSourceCleanupInvalid,
		)
	}
	return store.PersistSourceCleanupCompletion(ctx, SourceCleanupCompletion{
		Target: target, Receipt: deleted,
	})
}
