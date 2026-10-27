//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type httpIdempotencyPGFixture struct {
	owner   *sql.DB
	runtime *pgxpool.Pool
}

type pgxBoundIdempotencyDatabase struct {
	transaction pgx.Tx
}

func (database pgxBoundIdempotencyDatabase) Execute(
	ctx context.Context,
	statement string,
	arguments ...any,
) (platformpostgres.CommandResult, error) {
	_, err := database.transaction.Exec(ctx, statement, arguments...)
	return platformpostgres.CommandResult{}, err
}

func (database pgxBoundIdempotencyDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) platformpostgres.Row {
	return database.transaction.QueryRow(ctx, statement, arguments...)
}

func newHTTPIdempotencyPGFixture(
	t *testing.T,
	ctx context.Context,
	origin string,
) httpIdempotencyPGFixture {
	t.Helper()
	owner := startAsyncMigrationDatabase(t, ctx, "http_idempotency_"+origin)
	prepareAsyncAuthorityFixture(t, ctx, owner)
	addAsyncAuthorityStateColumns(t, ctx, owner)
	applyConsumerStoreMigrations(t, ctx, owner, origin)
	grantHTTPIdempotencyRuntime(t, ctx, owner)
	return httpIdempotencyPGFixture{
		owner: owner, runtime: openPublisherRuntimePool(t, ctx, owner),
	}
}

func grantHTTPIdempotencyRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const statement = `
CREATE ROLE harpy_runtime NOLOGIN
    NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_v2_runtime LOGIN PASSWORD 'runtime_integration_only'
    NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
GRANT harpy_runtime TO harpy_v2_runtime
    WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
GRANT USAGE ON SCHEMA platform TO harpy_runtime;
GRANT SELECT, INSERT ON platform.idempotency_receipts TO harpy_runtime;
GRANT INSERT ON platform.jobs TO harpy_runtime;
CREATE TABLE platform.http_idempotency_business_probe (
    id uuid PRIMARY KEY,
    value text NOT NULL
);
GRANT SELECT, INSERT ON platform.http_idempotency_business_probe
TO harpy_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("grant HTTP idempotency runtime: %v", err)
	}
}

func (fixture httpIdempotencyPGFixture) begin(
	t *testing.T,
	ctx context.Context,
	principal uuid.UUID,
) (pgx.Tx, *repository) {
	t.Helper()
	transaction, err := fixture.runtime.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	repository := &repository{
		database: pgxBoundIdempotencyDatabase{transaction: transaction},
		authority: authorityStamp{
			scopeID: uuid.MustParse(asyncTenantA), groupID: "operations",
			groupGeneration: 1, domain: "workflow", generation: 1,
		},
		principalID: principal, operationID: "forms.record.create",
	}
	return transaction, repository
}
