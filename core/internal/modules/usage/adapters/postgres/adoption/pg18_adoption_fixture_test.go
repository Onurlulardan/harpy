//go:build integration

package adoption

import (
	"context"
	"database/sql"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type sqlAdoptionDatabase struct{ transaction *sql.Tx }

func (database sqlAdoptionDatabase) QueryRow(ctx context.Context, statement string,
	arguments ...any,
) Row {
	return database.transaction.QueryRowContext(ctx, statement, arguments...)
}

func (database sqlAdoptionDatabase) Query(ctx context.Context, statement string,
	arguments ...any,
) (Rows, error) {
	rows, err := database.transaction.QueryContext(ctx, statement, arguments...)
	if err != nil {
		return nil, err
	}
	return sqlAdoptionRows{Rows: rows}, nil
}

type sqlAdoptionRows struct{ *sql.Rows }

func (rows sqlAdoptionRows) Close() { _ = rows.Rows.Close() }

func targetRunner(database *sql.DB) TransactionRunner {
	return func(ctx context.Context, action func(Database) error) error {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err = action(sqlAdoptionDatabase{transaction}); err != nil {
			_ = transaction.Rollback()
			return err
		}
		return transaction.Commit()
	}
}

func sourceRunner(database *sql.DB) LegacyTransactionRunner {
	return func(ctx context.Context, schema string,
		action func(LegacyDatabase) error,
	) error {
		transaction, err := database.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelRepeatableRead, ReadOnly: true})
		if err != nil {
			return err
		}
		if _, err = transaction.ExecContext(ctx,
			"SELECT pg_catalog.set_config('search_path',$1,true)",
			schema+",pg_catalog",
		); err != nil {
			_ = transaction.Rollback()
			return err
		}
		if err = action(sqlAdoptionDatabase{transaction}); err != nil {
			_ = transaction.Rollback()
			return err
		}
		return transaction.Commit()
	}
}

type integrationClock struct{ now time.Time }

func (clock integrationClock) Now() time.Time { return clock.now }

type crashTarget struct {
	usageapp.Target
	fail bool
}

func (target *crashTarget) Advance(ctx context.Context,
	command usageapp.ApplyPageCommand,
) (usageapp.Run, error) {
	if target.fail {
		target.fail = false
		return usageapp.Run{}, domain.ErrUnavailable
	}
	return target.Target.Advance(ctx, command)
}
