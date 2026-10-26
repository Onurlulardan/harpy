package postgres

import (
	"context"
	"reflect"
)

type TenancyCustomerProvisioningDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type tenancyCustomerProvisioningDatabase struct{ pool *Pool }

func (database tenancyCustomerProvisioningDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	if database.pool == nil || database.pool.closed.Load() {
		return &boundRow{err: ErrPoolClosed}
	}
	return database.pool.inner.QueryRow(ctx, statement, arguments...)
}

func BindTenancyCustomerProvisioning[T any](
	pool *Pool,
	factory func(TenancyCustomerProvisioningDatabase) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() {
		return zero, ErrPoolClosed
	}
	if !pool.runtimeCapable() || (pool.profile != ProfileAPI && pool.profile != ProfileWorker) {
		return zero, ErrTenancyPolicyProfile
	}
	if factory == nil {
		return zero, ErrTenancyPolicyBindingInvalid
	}
	database := tenancyCustomerProvisioningDatabase{pool: pool}
	result, err := factory(database)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(database) {
		return zero, ErrTenancyPolicyBindingInvalid
	}
	return result, nil
}
