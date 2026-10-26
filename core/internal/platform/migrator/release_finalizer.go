package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// FinalizationRow is the transaction-bound row surface exposed to release
// finalizers. Business adapters retain ownership of their SQL.
type FinalizationRow interface {
	Scan(destinations ...any) error
}

// FinalizationQueryRow executes one row-returning statement in the foundation
// finalization transaction.
type FinalizationQueryRow func(
	context.Context,
	string,
	...any,
) FinalizationRow

// ReleaseFinalizer persists release-bound module evidence before the database
// schema contract becomes ready.
type ReleaseFinalizer interface {
	FinalizeRelease(context.Context, FinalizationQueryRow) error
}

func FinalizationRowNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

type releaseFinalizationError struct {
	cause error
}

func (err *releaseFinalizationError) Error() string {
	return fmt.Sprintf("release finalization failed: %v", err.cause)
}

func (err *releaseFinalizationError) Unwrap() error {
	return err.cause
}

func (runner *Runner) finalizeRelease(
	ctx context.Context,
	queryRow FinalizationQueryRow,
) error {
	if runner.releaseFinalizer == nil {
		return nil
	}
	if queryRow == nil {
		return &releaseFinalizationError{
			cause: fmt.Errorf("finalization query dependency is incomplete"),
		}
	}
	if err := runner.releaseFinalizer.FinalizeRelease(ctx, queryRow); err != nil {
		return &releaseFinalizationError{cause: err}
	}
	return nil
}

func isReleaseFinalizationError(err error) bool {
	var target *releaseFinalizationError
	return errors.As(err, &target)
}
