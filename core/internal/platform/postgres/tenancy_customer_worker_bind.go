package postgres

import (
	"context"
	"reflect"
)

type TenancyCustomerWorkerDatabase interface {
	QueryRow(context.Context, string, ...any) Row
	Execute(context.Context, string, ...any) error
}

func (database tenancyCustomerWorkerDatabase) Execute(
	ctx context.Context,
	statement string,
	arguments ...any,
) error {
	if database.pool == nil || database.pool.closed.Load() {
		return ErrPoolClosed
	}
	_, err := database.pool.inner.Exec(ctx, statement, arguments...)
	return err
}

type tenancyCustomerWorkerDatabase struct{ pool *Pool }

func (database tenancyCustomerWorkerDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	if database.pool == nil || database.pool.closed.Load() {
		return &boundRow{err: ErrPoolClosed}
	}
	return database.pool.inner.QueryRow(ctx, statement, arguments...)
}

func BindTenancyCustomerWorker[T any](
	pool *Pool,
	factory func(TenancyCustomerWorkerDatabase) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() || pool.profile != ProfileWorker ||
		!pool.runtimeCapable() || factory == nil {
		return zero, ErrTenancyPolicyProfile
	}
	database := tenancyCustomerWorkerDatabase{pool: pool}
	result, err := factory(database)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(database) {
		return zero, ErrTenancyPolicyBindingInvalid
	}
	return result, nil
}
