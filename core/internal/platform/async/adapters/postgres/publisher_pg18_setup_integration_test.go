//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type publisherPGFixture struct {
	owner   *sql.DB
	runtime *pgxpool.Pool
	store   *workerRepository
	catalog eventsv1.ProducerCatalog
}

func newPublisherPGFixture(
	t *testing.T,
	ctx context.Context,
	origin string,
) publisherPGFixture {
	t.Helper()
	owner := startAsyncMigrationDatabase(t, ctx, "publisher_"+origin)
	applyAsyncACLFixture(t, ctx, owner, origin)
	addPublisherSecondTenant(t, ctx, owner)
	runtime := openPublisherRuntimePool(t, ctx, owner)
	return publisherPGFixture{
		owner: owner, runtime: runtime,
		store: &workerRepository{
			database: runtime, cleanup: 5 * time.Second,
			identity: &workerIdentity{marker: 1},
		},
		catalog: validPublisherCatalog(t),
	}
}

func addPublisherSecondTenant(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const statement = `
INSERT INTO platform.tenants VALUES ('22222222-2222-4222-8222-222222222222');
INSERT INTO platform.cutover_groups VALUES (
    'tenant','22222222-2222-4222-8222-222222222222','operations','v2','v2',1
);
INSERT INTO platform.domain_cutovers VALUES (
    'tenant','22222222-2222-4222-8222-222222222222','workflow','v2','v2',1,
    NULL,pg_catalog.clock_timestamp()
);
INSERT INTO platform.cutover_group_members VALUES (
    'tenant','22222222-2222-4222-8222-222222222222','operations','workflow'
);`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("add publisher second tenant: %v", err)
	}
}

func openPublisherRuntimePool(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) *pgxpool.Pool {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var connectionString string
	if err := connection.Raw(func(raw any) error {
		stdlibConnection, ok := raw.(*stdlib.Conn)
		if !ok {
			return fmt.Errorf("unexpected SQL driver connection %T", raw)
		}
		connectionString = stdlibConnection.Conn().Config().ConnString()
		return nil
	}); err != nil {
		t.Fatalf("resolve publisher database address: %v", err)
	}
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.User = "harpy_v2_runtime"
	config.ConnConfig.Password = "runtime_integration_only"
	config.AfterConnect = func(ctx context.Context, connection *pgx.Conn) error {
		_, err := connection.Exec(ctx, "SET ROLE harpy_runtime")
		return err
	}
	config.MaxConns = 8
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("open publisher runtime pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping publisher runtime pool: %v", err)
	}
	return pool
}
