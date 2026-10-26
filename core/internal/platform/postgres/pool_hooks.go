package postgres

import (
	"context"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func installPoolHooks(
	poolConfig *pgxpool.Config,
	settings config.Postgres,
	profile Profile,
	contract profileContract,
	binaryContract runtimev1.BinarySchemaContract,
	collector *startupCollector,
) {
	poolConfig.AfterConnect = func(ctx context.Context, connection *pgx.Conn) error {
		result, err := validateStartup(ctx, connection, profile, contract, binaryContract)
		if err != nil {
			return err
		}
		return collector.accept(result)
	}
	poolConfig.PrepareConn = func(ctx context.Context, connection *pgx.Conn) (bool, error) {
		return reusableSession(ctx, connection, contract)
	}
	poolConfig.AfterRelease = func(connection *pgx.Conn) bool {
		ctx, cancel := context.WithTimeout(context.Background(), settings.CleanupTimeout())
		defer cancel()
		reusable, _ := reusableSession(ctx, connection, contract)
		return reusable
	}
}
