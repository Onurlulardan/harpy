//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	postgresadoption "github.com/Onurlulardan/harpy/core/internal/modules/support/adapters/postgres/adoption"
	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

type adoptionTx struct{ transaction *sql.Tx }

func (database adoptionTx) QueryRow(ctx context.Context, query string,
	arguments ...any,
) postgresadoption.Row {
	return database.transaction.QueryRowContext(ctx, query, arguments...)
}

func adopterStore(t *testing.T, ctx context.Context, database *sql.DB) (
	*postgresadoption.Store, *sql.Conn,
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(ctx, `SET SESSION AUTHORIZATION harpy_adopter`); err != nil {
		connection.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = connection.ExecContext(context.Background(), `RESET SESSION AUTHORIZATION`)
		_ = connection.Close()
	})
	runner := func(ctx context.Context, _ postgresadoption.TransactionMode,
		execute func(postgresadoption.Database) error,
	) error {
		transaction, beginErr := connection.BeginTx(ctx, nil)
		if beginErr != nil {
			return beginErr
		}
		if executeErr := execute(adoptionTx{transaction: transaction}); executeErr != nil {
			_ = transaction.Rollback()
			return executeErr
		}
		return transaction.Commit()
	}
	store, err := postgresadoption.NewStoreForAdopter(runner)
	if err != nil {
		t.Fatal(err)
	}
	return store, connection
}

type crashBeforeAdvance struct {
	application.Target
	failed bool
}

func (target *crashBeforeAdvance) AdoptionRecordAdvance(context.Context,
	application.ApplyPageCommand,
) (application.Run, error) {
	if !target.failed {
		target.failed = true
		return application.Run{}, application.ErrUnavailable
	}
	return application.Run{}, errors.New("unexpected second crash target advance")
}
