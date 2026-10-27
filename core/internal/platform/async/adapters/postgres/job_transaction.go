package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func withJobTransaction[T any](
	ctx context.Context,
	repository *workerRepository,
	claim JobClaim,
	operation func(pgx.Tx) (T, error),
) (T, error) {
	var zero T
	if ctx == nil {
		return zero, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return zero, ErrWorkerStoreUnavailable
	}
	if !claim.valid(repository.identity) {
		return zero, ErrJobClaimInvalid
	}
	return withJobAuthorityTransaction(ctx, repository, claim.authority, operation)
}
