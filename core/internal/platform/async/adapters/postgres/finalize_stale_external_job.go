package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) finalizeStaleExternalJob(
	ctx context.Context,
	job staleExternalJob,
) (finalized bool, resultErr error) {
	if ctx == nil {
		return false, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return false, ErrWorkerStoreUnavailable
	}
	if !job.valid(repository.identity) {
		return false, ErrStaleJobInvalid
	}
	transaction, err := repository.database.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return false, fmt.Errorf("begin stale external finalization: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			resultErr = errors.Join(
				resultErr, rollbackJobClaim(transaction, repository.cleanup),
			)
		}
	}()
	if err := transaction.QueryRow(
		ctx, finalizeStaleExternalJobSQL, job.arguments()...,
	).Scan(&finalized); err != nil {
		return false, fmt.Errorf("finalize stale external job: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit stale external finalization: %w", err)
	}
	committed = true
	return finalized, nil
}
