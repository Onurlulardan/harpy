package systemtransfer

import (
	"context"
	"errors"
	"time"
)

const importSourceCleanupRetryMaximum = 15 * time.Minute

func RetrySourceCleanup(ctx context.Context,
	store SourceCleanupStore,
	clock SourceCleanupClock,
	jitter SourceCleanupJitter,
	target SourceCleanupTarget,
	cause error,
) (bool, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(store) ||
		nilDependency(clock) || nilDependency(jitter) ||
		target.Validate() != nil || cause == nil {
		return false, ErrSourceCleanupInvalid
	}
	observedAt := clock.Now().UTC()
	if !controlTime(observedAt) {
		return false, ErrSourceCleanupInvalid
	}
	limit := sourceCleanupRetryLimit(target.Attempt)
	delay := time.Duration(jitter.Uint64N(uint64(limit) + 1))
	code := SourceCleanupUnavailable
	if errors.Is(cause, ErrSourceDeleteOutcomeUnknown) {
		code = SourceCleanupOutcomeUnknown
	}
	return store.ScheduleSourceCleanupRetry(ctx, SourceCleanupRetry{
		Target: target, ErrorCode: code, ObservedAt: observedAt,
		RetryAt: observedAt.Add(delay),
	})
}

func sourceCleanupRetryLimit(attempt int) time.Duration {
	if attempt < 1 {
		return 0
	}
	limit := 5 * time.Second
	for current := 1; current < attempt && limit < importSourceCleanupRetryMaximum; current++ {
		limit *= 2
		if limit > importSourceCleanupRetryMaximum {
			return importSourceCleanupRetryMaximum
		}
	}
	return limit
}
