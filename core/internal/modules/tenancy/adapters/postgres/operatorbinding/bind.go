package operatorbinding

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

type Database interface {
	QueryRow(context.Context, string, ...any) Row
	Run(context.Context, string, ...any) error
}

type database struct {
	bound platformpostgres.TenancyOperatorDatabase
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

func Bind[T any](pool Pool, factory func(Database) (T, error)) (T, error) {
	var zero T
	if pool == nil || factory == nil {
		return zero, platformpostgres.ErrTenancyPolicyBindingInvalid
	}
	return platformpostgres.BindTenancyOperator(
		pool.PostgresOwnerBindingPool(),
		func(bound platformpostgres.TenancyOperatorDatabase) (T, error) {
			return factory(database{bound: bound})
		},
	)
}
