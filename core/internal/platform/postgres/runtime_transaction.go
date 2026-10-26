package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
)

// RuntimeTransaction is the single PostgreSQL transaction foundation used by
// tenant, global, provider-target, and job-target scopes.
type RuntimeTransaction struct {
	transaction pgx.Tx
	open        atomic.Bool
	bindings    *bindingState
	cleanup     time.Duration
}

func (pool *Pool) BeginRuntimeTransaction(
	ctx context.Context,
	mode TransactionMode,
) (*RuntimeTransaction, error) {
	if !pool.runtimeCapable() {
		return nil, ErrRuntimeProfile
	}
	return pool.beginProfileTransaction(ctx, mode)
}

func (pool *Pool) beginProfileTransaction(
	ctx context.Context,
	mode TransactionMode,
) (*RuntimeTransaction, error) {
	if !pool.profileTransactionCapable() {
		return nil, ErrRuntimeProfile
	}
	options, valid := mode.options()
	if !valid {
		return nil, fmt.Errorf("postgres runtime transaction mode is invalid")
	}
	transaction, err := pool.inner.BeginTx(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("postgres runtime transaction begin failed")
	}
	result := &RuntimeTransaction{
		transaction: transaction, bindings: newBindingState(), cleanup: pool.cleanupTimeout,
	}
	result.open.Store(true)
	return result, nil
}

func (transaction *RuntimeTransaction) SealBindings() error {
	if transaction == nil || transaction.bindings == nil {
		return ErrTransactionClosed
	}
	return transaction.bindings.seal(transaction.cleanup)
}

func (transaction *RuntimeTransaction) Commit(ctx context.Context) error {
	if transaction == nil || !transaction.open.CompareAndSwap(true, false) {
		return ErrTransactionClosed
	}
	if err := transaction.SealBindings(); err != nil {
		_ = transaction.destroyTainted()
		return err
	}
	err := transaction.transaction.Commit(ctx)
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrTxCommitRollback) {
		return ErrCommitRolledBack
	}
	return ErrCommitAmbiguous
}

func (transaction *RuntimeTransaction) Rollback() error {
	if transaction == nil || !transaction.open.CompareAndSwap(true, false) {
		return ErrTransactionClosed
	}
	if err := transaction.SealBindings(); err != nil {
		return transaction.destroyTainted()
	}
	ctx, cancel := context.WithTimeout(context.Background(), transaction.cleanup)
	defer cancel()
	return transaction.transaction.Rollback(ctx)
}

func (transaction *RuntimeTransaction) destroyTainted() error {
	ctx, cancel := context.WithTimeout(context.Background(), transaction.cleanup)
	defer cancel()
	closeErr := transaction.transaction.Conn().Close(ctx)
	rollbackErr := transaction.transaction.Rollback(ctx)
	if closeErr != nil {
		return closeErr
	}
	if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
		return rollbackErr
	}
	return nil
}

func (transaction *RuntimeTransaction) usable() bool {
	return transaction != nil && transaction.open.Load()
}
