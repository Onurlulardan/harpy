package postgres

import (
	"context"
	"reflect"
)

type IAMInitialAdminSetupDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type iamInitialAdminSetupDatabase struct{ pool *Pool }

func (database iamInitialAdminSetupDatabase) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) Row {
	if database.pool == nil || database.pool.closed.Load() {
		return &boundRow{err: ErrPoolClosed}
	}
	return database.pool.inner.QueryRow(ctx, statement, arguments...)
}

func BindIAMInitialAdminSetup[T any](
	pool *Pool,
	factory func(IAMInitialAdminSetupDatabase) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() || pool.profile != ProfileAPI ||
		!pool.runtimeCapable() || factory == nil {
		return zero, ErrAPIStoreProfile
	}
	database := iamInitialAdminSetupDatabase{pool: pool}
	result, err := factory(database)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(database) {
		return zero, ErrAPIStoreBindingInvalid
	}
	return result, nil
}
