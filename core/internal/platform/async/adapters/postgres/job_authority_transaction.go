package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func withJobAuthorityTransaction[T any](
	ctx context.Context,
	repository *workerRepository,
	authority JobAuthority,
	operation func(pgx.Tx) (T, error),
) (result T, resultErr error) {
	if ctx == nil {
		return result, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return result, ErrWorkerStoreUnavailable
	}
	if !authority.valid(repository.identity) {
		return result, ErrJobAuthorityInvalid
	}
	transaction, err := repository.database.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return result, fmt.Errorf("begin job authority transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			resultErr = errors.Join(
				resultErr, rollbackJobClaim(transaction, repository.cleanup),
			)
		}
	}()
	if _, err := transaction.Exec(
		ctx, assertJobAuthoritySQL, authorityAssertionArguments(authority)...,
	); err != nil {
		return result, classifyJobAuthorityError(err)
	}
	result, err = operation(transaction)
	if err != nil {
		return result, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit job authority transaction: %w", err)
	}
	committed = true
	return result, nil
}
