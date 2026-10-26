package postgres

import (
	"context"
	"reflect"
)

type TenancyOperatorDatabase interface {
	QueryRow(context.Context, string, ...any) Row
	Execute(context.Context, string, ...any) error
}

type tenancyOperatorDatabase struct{ pool *Pool }

func (database tenancyOperatorDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	if database.pool == nil || database.pool.closed.Load() {
		return &boundRow{err: ErrPoolClosed}
	}
	return database.pool.inner.QueryRow(ctx, statement, arguments...)
}

func (database tenancyOperatorDatabase) Execute(
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

func BindTenancyOperator[T any](
	pool *Pool,
	factory func(TenancyOperatorDatabase) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() {
		return zero, ErrPoolClosed
	}
	if !pool.runtimeCapable() || pool.profile != ProfileOperatorInitialize {
		return zero, ErrTenancyPolicyProfile
	}
	if factory == nil {
		return zero, ErrTenancyPolicyBindingInvalid
	}
	database := tenancyOperatorDatabase{pool: pool}
	result, err := factory(database)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(database) {
		return zero, ErrTenancyPolicyBindingInvalid
	}
	return result, nil
}
