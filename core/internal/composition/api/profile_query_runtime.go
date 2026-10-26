package api

import (
	"context"
	"fmt"
	"time"

	supportmodule "github.com/Onurlulardan/harpy/core/internal/modules/support"
	usagemodule "github.com/Onurlulardan/harpy/core/internal/modules/usage"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type profileQueryRuntime struct {
	pool    *pgxpool.Pool
	cleanup time.Duration
}

type usageProfileDatabase struct{ transaction pgx.Tx }

func (database usageProfileDatabase) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) usagemodule.PostgresRow {
	return database.transaction.QueryRow(ctx, statement, arguments...)
}

type supportProfileDatabase struct{ transaction pgx.Tx }

func (database supportProfileDatabase) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) supportmodule.RuntimeRow {
	return database.transaction.QueryRow(ctx, statement, arguments...)
}

func newProfileQueryRuntime(
	pool *platformpostgres.Pool,
) (*profileQueryRuntime, error) {
	runtime, err := platformpostgres.BindAPIStore(pool,
		func(raw *pgxpool.Pool, cleanup time.Duration) (*profileQueryRuntime, error) {
			if raw == nil || cleanup <= 0 {
				return nil, fmt.Errorf("profile query runtime dependencies are invalid")
			}
			return &profileQueryRuntime{pool: raw, cleanup: cleanup}, nil
		})
	if err != nil {
		return nil, fmt.Errorf("bind profile query runtime: %w", err)
	}
	return runtime, nil
}

func (runtime *profileQueryRuntime) usageRunner() usagemodule.PostgresTransactionRunner {
	return func(ctx context.Context, mode usagemodule.PostgresTransactionMode,
		execute func(usagemodule.PostgresDatabase) error,
	) error {
		options, valid := profileQueryOptions(string(mode))
		if !valid || execute == nil {
			return fmt.Errorf("usage profile transaction is invalid")
		}
		return runtime.within(ctx, options, func(transaction pgx.Tx) error {
			return execute(usageProfileDatabase{transaction: transaction})
		})
	}
}

func (runtime *profileQueryRuntime) supportRunner() supportmodule.RuntimeTransactionRunner {
	return func(ctx context.Context, mode supportmodule.RuntimeTransactionMode,
		execute func(supportmodule.RuntimeDatabase) error,
	) error {
		options, valid := profileQueryOptions(string(mode))
		if !valid || execute == nil {
			return fmt.Errorf("support profile transaction is invalid")
		}
		return runtime.within(ctx, options, func(transaction pgx.Tx) error {
			return execute(supportProfileDatabase{transaction: transaction})
		})
	}
}

func profileQueryOptions(mode string) (pgx.TxOptions, bool) {
	switch mode {
	case "command":
		return pgx.TxOptions{IsoLevel: pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite, DeferrableMode: pgx.NotDeferrable}, true
	case "snapshot":
		return pgx.TxOptions{IsoLevel: pgx.RepeatableRead,
			AccessMode: pgx.ReadOnly, DeferrableMode: pgx.NotDeferrable}, true
	default:
		return pgx.TxOptions{}, false
	}
}

func (runtime *profileQueryRuntime) within(
	ctx context.Context, options pgx.TxOptions, execute func(pgx.Tx) error,
) error {
	if runtime == nil || runtime.pool == nil || runtime.cleanup <= 0 ||
		ctx == nil || ctx.Err() != nil || execute == nil {
		return fmt.Errorf("profile query runtime is unavailable")
	}
	transaction, err := runtime.pool.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("begin profile transaction: %w", err)
	}
	defer func() {
		rollbackContext, cancel := context.WithTimeout(context.Background(), runtime.cleanup)
		defer cancel()
		_ = transaction.Rollback(rollbackContext)
	}()
	if err = execute(transaction); err != nil {
		return err
	}
	if err = transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit profile transaction: %w", err)
	}
	return nil
}
