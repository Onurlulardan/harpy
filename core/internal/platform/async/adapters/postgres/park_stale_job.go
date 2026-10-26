package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) ParkStaleJob(
	ctx context.Context,
	claim JobClaim,
) (parked bool, resultErr error) {
	if ctx == nil {
		return false, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return false, ErrWorkerStoreUnavailable
	}
	job, err := staleJobFromClaim(repository.identity, claim)
	if err != nil {
		return false, err
	}
	return repository.parkStaleJob(ctx, job)
}

func (repository *workerRepository) parkStaleJob(
	ctx context.Context,
	job staleJob,
) (parked bool, resultErr error) {
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
		return false, fmt.Errorf("begin stale job park transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			resultErr = errors.Join(
				resultErr, rollbackJobClaim(transaction, repository.cleanup),
			)
		}
	}()
	err = transaction.QueryRow(
		ctx, parkStaleJobSQL,
		job.jobID, job.state, job.leaseOwnerArgument(),
		job.leaseGeneration, uuid.New(),
	).Scan(&parked)
	if err != nil {
		return false, fmt.Errorf("park stale job: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit stale job park: %w", err)
	}
	committed = true
	return parked, nil
}
