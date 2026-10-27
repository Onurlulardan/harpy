package postgres

import (
	"context"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var asyncWorkerRawPoolType = reflect.TypeOf((*pgxpool.Pool)(nil))

type AsyncWorkerExecutor interface {
	QueryRow(context.Context, string, ...any) Row
}

type asyncWorkerExecutor struct{ database *pgxpool.Pool }

func (executor asyncWorkerExecutor) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	return executor.database.QueryRow(ctx, statement, arguments...)
}

// BindAsyncWorkerStore grants the async owner adapter a worker-only store factory.
func BindAsyncWorkerStore[T any](
	pool *Pool,
	factory func(*pgxpool.Pool, time.Duration) T,
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() {
		return zero, ErrPoolClosed
	}
	if pool.profile != ProfileWorker {
		return zero, ErrAsyncWorkerProfile
	}
	if factory == nil {
		return zero, ErrAsyncWorkerBindingInvalid
	}
	result := factory(pool.inner, pool.cleanupTimeout)
	if reflect.TypeOf(any(result)) == asyncWorkerRawPoolType {
		return zero, ErrAsyncWorkerBindingInvalid
	}
	return result, nil
}

func BindAsyncWorkerExecutor[T any](
	pool *Pool,
	factory func(AsyncWorkerExecutor) T,
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() {
		return zero, ErrPoolClosed
	}
	if pool.profile != ProfileWorker {
		return zero, ErrAsyncWorkerProfile
	}
	if factory == nil {
		return zero, ErrAsyncWorkerBindingInvalid
	}
	result := factory(asyncWorkerExecutor{database: pool.inner})
	if reflect.TypeOf(any(result)) == reflect.TypeOf(asyncWorkerExecutor{}) {
		return zero, ErrAsyncWorkerBindingInvalid
	}
	return result, nil
}
