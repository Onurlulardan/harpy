package systemtransfer

import (
	"context"
	"errors"
)

func (runner *SourceCleanupRunner) execute(ctx context.Context,
	target SourceCleanupTarget,
) error {
	deleted, err := runner.objects.DeleteExactImportSource(
		ctx, target.ObjectKey, target.ObjectVersion,
	)
	if err == nil {
		_, err = FinalizeSourceCleanup(ctx, runner.store, target, deleted)
		if err == nil || errors.Is(err, ErrSourceCleanupStale) {
			return nil
		}
		if !errors.Is(err, ErrSourceDeleteOutcomeUnknown) {
			return err
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	_, retryErr := RetrySourceCleanup(
		ctx, runner.store, runner.clock, runner.jitter, target, err,
	)
	if errors.Is(retryErr, ErrSourceCleanupStale) {
		return nil
	}
	return retryErr
}
