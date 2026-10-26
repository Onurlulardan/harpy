package postgres

import (
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var apiRawPoolType = reflect.TypeOf((*pgxpool.Pool)(nil))

// BindAPIStore grants composition an API-profile-only adapter factory.
func BindAPIStore[T any](
	pool *Pool,
	factory func(*pgxpool.Pool, time.Duration) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() {
		return zero, ErrPoolClosed
	}
	if pool.profile != ProfileAPI || !pool.runtimeCapable() {
		return zero, ErrAPIStoreProfile
	}
	if factory == nil {
		return zero, ErrAPIStoreBindingInvalid
	}
	result, err := factory(pool.inner, pool.cleanupTimeout)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == apiRawPoolType {
		return zero, ErrAPIStoreBindingInvalid
	}
	return result, nil
}
