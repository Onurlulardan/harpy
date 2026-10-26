package postgres

import (
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var workerRawPoolType = reflect.TypeOf((*pgxpool.Pool)(nil))

// BindWorkerStore grants composition a worker-profile adapter factory.
func BindWorkerStore[T any](
	pool *Pool,
	factory func(*pgxpool.Pool, time.Duration) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() {
		return zero, ErrPoolClosed
	}
	if pool.profile != ProfileWorker || !pool.runtimeCapable() {
		return zero, ErrWorkerStoreProfile
	}
	if factory == nil {
		return zero, ErrWorkerStoreBindingInvalid
	}
	result, err := factory(pool.inner, pool.cleanupTimeout)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == workerRawPoolType {
		return zero, ErrWorkerStoreBindingInvalid
	}
	return result, nil
}
