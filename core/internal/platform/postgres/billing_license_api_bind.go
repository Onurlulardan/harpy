package postgres

import (
	"context"
	"reflect"
)

type BillingLicenseAPIDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type billingLicenseAPIDatabase struct{ pool *Pool }

func (database billingLicenseAPIDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	if database.pool == nil || database.pool.closed.Load() {
		return &boundRow{err: ErrPoolClosed}
	}
	return database.pool.inner.QueryRow(ctx, statement, arguments...)
}

func BindBillingLicenseAPI[T any](
	pool *Pool,
	factory func(BillingLicenseAPIDatabase) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() || pool.profile != ProfileAPI ||
		!pool.runtimeCapable() || factory == nil {
		return zero, ErrAPIStoreProfile
	}
	database := billingLicenseAPIDatabase{pool: pool}
	result, err := factory(database)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(database) {
		return zero, ErrAPIStoreBindingInvalid
	}
	return result, nil
}
