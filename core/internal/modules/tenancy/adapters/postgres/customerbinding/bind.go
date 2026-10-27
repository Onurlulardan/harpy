package customerbinding

import (
	"context"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type Pool interface {
	PostgresOwnerBindingPool() *platformpostgres.Pool
}

type Row interface {
	Scan(...any) error
}

type WorkerDatabase interface {
	QueryRow(context.Context, string, ...any) Row
	Run(context.Context, string, ...any) error
}

type ProvisioningDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type database struct {
	bound platformpostgres.TenancyCustomerWorkerDatabase
}

func (value database) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	return value.bound.QueryRow(ctx, statement, arguments...)
}

func (value database) Run(
	ctx context.Context,
	statement string,
	arguments ...any,
) error {
	return value.bound.Execute(ctx, statement, arguments...)
}

type provisioningDatabase struct {
	bound platformpostgres.TenancyCustomerProvisioningDatabase
}

func (value provisioningDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	return value.bound.QueryRow(ctx, statement, arguments...)
}

func WithCustomerWorker[T any](
	pool Pool,
	factory func(WorkerDatabase) (T, error),
) (T, error) {
	return platformpostgres.BindTenancyCustomerWorker(
		pool.PostgresOwnerBindingPool(),
		func(bound platformpostgres.TenancyCustomerWorkerDatabase) (T, error) {
			return factory(database{bound: bound})
		},
	)
}

func WithCustomerProvisioning[T any](
	pool Pool,
	factory func(ProvisioningDatabase) (T, error),
) (T, error) {
	return platformpostgres.BindTenancyCustomerProvisioning(
		pool.PostgresOwnerBindingPool(),
		func(bound platformpostgres.TenancyCustomerProvisioningDatabase) (T, error) {
			return factory(provisioningDatabase{bound: bound})
		},
	)
}

func RowNotFound(err error) bool {
	return platformpostgres.RowNotFound(err)
}
