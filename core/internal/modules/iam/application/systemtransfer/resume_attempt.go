package systemtransfer

import (
	"context"
	"errors"
)

func (runtime *ImportRuntime) ResumeAttempt(
	ctx context.Context,
	claim AttemptClaim,
) (attempt Attempt, resultErr error) {
	if ctx == nil || ctx.Err() != nil || runtime == nil ||
		nilDependency(runtime.loader) || nilDependency(runtime.sources) ||
		runtime.batches == nil || claim.Validate() != nil {
		return Attempt{}, ErrInvalidAttempt
	}
	state, err := runtime.loader.LoadAttempt(ctx, claim)
	if err != nil {
		return Attempt{}, workerResult(ctx, err)
	}
	attempt = state.Attempt
	defer func() { resultErr = attachWorkerActor(resultErr, state.Attempt.ActorID) }()
	if state.Validate(claim) != nil {
		return attempt, classifyWorkerError(ErrWorkerState)
	}
	if state.Finished {
		return attempt, nil
	}
	plaintext, err := runtime.sources.OpenExactImportSource(ctx, state.Source)
	if err != nil || plaintext == nil {
		return attempt, classifyWorkerError(ErrWorkerUnavailable)
	}
	defer func() {
		if closeErr := plaintext.Close(); closeErr != nil && resultErr == nil {
			resultErr = classifyWorkerError(ErrWorkerUnavailable)
		}
	}()
	if err = runtime.resumeSource(ctx, state, plaintext); err != nil {
		return attempt, workerResult(ctx, err)
	}
	return attempt, nil
}

func workerResult(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return classifyWorkerError(err)
}
