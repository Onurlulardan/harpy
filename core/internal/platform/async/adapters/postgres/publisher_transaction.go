package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func withPublisherTransaction[T any](
	ctx context.Context,
	repository *workerRepository,
	authority OutboxAuthority,
	operation func(pgx.Tx) (T, error),
) (result T, resultErr error) {
	if ctx == nil {
		return result, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return result, ErrWorkerStoreUnavailable
	}
	if !authority.valid(repository.identity) {
		return result, ErrOutboxAuthorityInvalid
	}
	transaction, err := repository.database.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return result, fmt.Errorf("begin publisher transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			resultErr = errors.Join(
				resultErr, rollbackPublisherTransaction(transaction, repository.cleanup),
			)
		}
	}()
	if _, err := transaction.Exec(
		ctx, assertJobAuthoritySQL, outboxAuthorityArguments(authority)...,
	); err != nil {
		return result, classifyOutboxAuthorityError(err)
	}
	result, err = operation(transaction)
	if err != nil {
		return result, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit publisher transaction: %w", err)
	}
	committed = true
	return result, nil
}

func classifyOutboxAuthorityError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "P5601", "P5602", "P5603":
			return fmt.Errorf("%w: database authority assertion failed", ErrOutboxAuthorityStale)
		}
	}
	return fmt.Errorf("assert current outbox authority: %w", err)
}
