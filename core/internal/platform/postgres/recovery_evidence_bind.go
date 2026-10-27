package postgres

import (
	"context"
	"fmt"
	"reflect"
)

const iamRecoveryEvidenceOwner = "iam"

type RecoveryEvidenceDatabase interface {
	Query(context.Context, string, ...any) (Rows, error)
	QueryRow(context.Context, string, ...any) Row
}

type RecoveryEvidenceTransactions interface {
	Within(
		context.Context,
		TransactionMode,
		func(RecoveryEvidenceDatabase) error,
	) error
}

type recoveryEvidenceDatabase struct{ executor BoundExecutor }
type recoveryEvidenceTransactions struct {
	pool  *Pool
	owner string
}

func (database recoveryEvidenceDatabase) Query(
	ctx context.Context,
	statement string,
	arguments ...any,
) (Rows, error) {
	return database.executor.Query(ctx, statement, arguments...)
}

func (database recoveryEvidenceDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	return database.executor.QueryRow(ctx, statement, arguments...)
}

func (runner recoveryEvidenceTransactions) Within(
	ctx context.Context,
	mode TransactionMode,
	callback func(RecoveryEvidenceDatabase) error,
) error {
	if callback == nil || runner.pool == nil ||
		runner.pool.profile != ProfileDatabaseBackupExecutor {
		return ErrOwnerBindingInvalid
	}
	transaction, err := runner.pool.beginProfileTransaction(ctx, mode)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	binder, err := NewOwnerBinder(
		runner.owner,
		func(executor BoundExecutor) recoveryEvidenceDatabase {
			return recoveryEvidenceDatabase{executor: executor}
		},
	)
	if err != nil {
		return err
	}
	database, err := BindOwner(transaction, binder)
	if err != nil {
		return err
	}
	if err = callback(database); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

func BindIAMRecoveryEvidenceStore[T any](
	pool *Pool,
	factory func(RecoveryEvidenceTransactions) (T, error),
) (T, error) {
	return bindRecoveryEvidenceStore(pool, iamRecoveryEvidenceOwner, factory)
}

func BindPlatformAuthorityRecoveryEvidenceStore[T any](
	pool *Pool,
	factory func(RecoveryEvidenceTransactions) (T, error),
) (T, error) {
	return bindRecoveryEvidenceStore(pool, platformAuthorityOwner, factory)
}

func bindRecoveryEvidenceStore[T any](
	pool *Pool,
	owner string,
	factory func(RecoveryEvidenceTransactions) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() ||
		pool.profile != ProfileDatabaseBackupExecutor ||
		!pool.profileTransactionCapable() || factory == nil {
		return zero, fmt.Errorf("recovery evidence binding is unavailable")
	}
	runner := recoveryEvidenceTransactions{pool: pool, owner: owner}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
