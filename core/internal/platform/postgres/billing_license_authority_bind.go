package postgres

import (
	"context"
	"reflect"
)

type BillingLicenseAuthorityDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type billingLicenseAuthorityDatabase struct{ pool *Pool }

func (database billingLicenseAuthorityDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	if database.pool == nil || database.pool.closed.Load() {
		return &boundRow{err: ErrPoolClosed}
	}
	return database.pool.inner.QueryRow(ctx, statement, arguments...)
}

func BindBillingLicenseAuthority[T any](
	pool *Pool,
	factory func(BillingLicenseAuthorityDatabase) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() || !pool.runtimeCapable() ||
		!licenseAuthorityProfile(pool.profile) || factory == nil {
		return zero, ErrAPIStoreProfile
	}
	database := billingLicenseAuthorityDatabase{pool: pool}
	result, err := factory(database)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(database) {
		return zero, ErrAPIStoreBindingInvalid
	}
	return result, nil
}

func licenseAuthorityProfile(profile Profile) bool {
	return profile == ProfileWorker ||
		profile == ProfileOperatorLicenseReconcile
}
