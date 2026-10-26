package postgres

import (
	"context"
	"reflect"
)

type TenancyPolicyQueryRow interface {
	QueryRow(context.Context, string, ...any) Row
}

type tenancyPolicyQueryRow struct{ pool *Pool }

func (query tenancyPolicyQueryRow) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	if query.pool == nil || query.pool.closed.Load() {
		return &boundRow{err: ErrPoolClosed}
	}
	return query.pool.inner.QueryRow(ctx, statement, arguments...)
}

func BindTenancyPolicyReader[T any](
	pool *Pool,
	factory func(TenancyPolicyQueryRow) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() {
		return zero, ErrPoolClosed
	}
	if !pool.runtimeCapable() ||
		(pool.profile != ProfileAPI && pool.profile != ProfileWorker) {
		return zero, ErrTenancyPolicyProfile
	}
	if factory == nil {
		return zero, ErrTenancyPolicyBindingInvalid
	}
	query := tenancyPolicyQueryRow{pool: pool}
	result, err := factory(query)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(query) {
		return zero, ErrTenancyPolicyBindingInvalid
	}
	return result, nil
}
