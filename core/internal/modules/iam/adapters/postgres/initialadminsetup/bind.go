package initialadminsetup

import (
	"context"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type RuntimePool interface {
	PostgresOwnerBindingPool() *platformpostgres.Pool
}

func NewRuntimeStore(handle RuntimePool) (*Store, error) {
	if handle == nil {
		return nil, ErrUnavailable
	}
	pool := handle.PostgresOwnerBindingPool()
	return platformpostgres.BindIAMInitialAdminSetup(
		pool,
		func(database platformpostgres.IAMInitialAdminSetupDatabase) (*Store, error) {
			return NewStore(queryRow(database.QueryRow))
		},
	)
}

func RunReissue(
	ctx context.Context,
	handle RuntimePool,
	schemaName string,
	callback func(QueryRow) error,
) error {
	if handle == nil || callback == nil {
		return ErrUnavailable
	}
	return platformpostgres.RunIAMInitialAdminSetupReissue(
		ctx,
		handle.PostgresOwnerBindingPool(),
		schemaName,
		func(database platformpostgres.IAMInitialAdminSetupReissueDatabase) error {
			return callback(queryRow(database.QueryRow))
		},
	)
}

func queryRow(
	run func(context.Context, string, ...any) platformpostgres.Row,
) QueryRow {
	return func(ctx context.Context, statement string, arguments ...any) Row {
		return run(ctx, statement, arguments...)
	}
}
