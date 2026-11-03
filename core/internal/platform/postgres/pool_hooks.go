package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func installPoolHooks(
	poolConfig *pgxpool.Config,
	settings config.Postgres,
	profile Profile,
	contract profileContract,
	collector *startupCollector,
) {
	poolConfig.AfterConnect = func(ctx context.Context, connection *pgx.Conn) error {
		result, err := validateStartup(ctx, connection, profile, contract)
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
