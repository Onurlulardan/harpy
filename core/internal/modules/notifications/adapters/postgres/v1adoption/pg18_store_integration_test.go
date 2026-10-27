//go:build integration

package v1adoption

import (
	"context"
	"errors"
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func transactionAdoptionStore(transaction pgx.Tx, tenantID uuid.UUID) *Store {
	return &Store{tenantID: tenantID,
		queryRow: func(ctx context.Context, query string, arguments ...any) platformpostgres.Row {
			return transaction.QueryRow(ctx, query, arguments...)
		},
		query: func(ctx context.Context, query string, arguments ...any) (platformpostgres.Rows, error) {
			return transaction.Query(ctx, query, arguments...)
		},
		execute: func(ctx context.Context, query string, arguments ...any) error {
			_, err := transaction.Exec(ctx, query, arguments...)
			return err
		}, notFound: func(err error) bool { return errors.Is(err, pgx.ErrNoRows) },
		allowed: func(context.Context) bool { return true }}
}

func executeNotificationStep(t *testing.T, ctx context.Context, connection *pgx.Conn,
	schema string, plan app.Plan, commit bool,
) app.StepResult {
	t.Helper()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.Exec(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
		schema+", pg_catalog"); err != nil {
		_ = transaction.Rollback(ctx)
		t.Fatal(err)
	}
	runner, err := app.NewRunner(transactionAdoptionStore(transaction, plan.TenantID))
	if err != nil {
		_ = transaction.Rollback(ctx)
		t.Fatal(err)
	}
	result, err := runner.Execute(ctx, plan)
	if err != nil {
		_ = transaction.Rollback(ctx)
		t.Fatalf("adoption step: %v", err)
	}
	if commit {
		err = transaction.Commit(ctx)
	} else {
		err = transaction.Rollback(ctx)
	}
	if err != nil {
		t.Fatal(err)
	}
	return result
}
