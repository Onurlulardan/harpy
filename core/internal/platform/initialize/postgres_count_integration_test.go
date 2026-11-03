//go:build integration

package initialize

import (
	"context"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5"
)

func openAdmin(t *testing.T) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(context.Background(), adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func openDatabase(t *testing.T, dsn, database string) *pgx.Conn {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/" + database
	conn, err := pgx.Connect(context.Background(), parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func relationCount(t *testing.T, database string) int {
	t.Helper()
	conn := openDatabase(t, adminDSN, database)
	var count int
	err := conn.QueryRow(context.Background(), `SELECT count(*) FROM pg_class relation
JOIN pg_namespace namespace ON namespace.oid = relation.relnamespace
WHERE namespace.nspname NOT IN ('pg_catalog', 'information_schema')
  AND namespace.nspname NOT LIKE 'pg\_%' ESCAPE '\'
  AND relation.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')`).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}
