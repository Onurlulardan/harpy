//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const rolledBackLayoutRevision = "eeeeeeee-eeee-7eee-8eee-eeeeeeeeeeee"

func assertLayoutRollback(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction, err := beginLayoutRuntimeTx(ctx, database, "layout_a")
	if err != nil {
		t.Fatal(err)
	}
	var document string
	err = transaction.QueryRowContext(ctx, createSQL,
		rolledBackLayoutRevision, string(layoutapi.ResourceRole), adapterDefinition,
		layoutDigest(adapterDefinition), layoutNow(8)).Scan(&document)
	if err != nil {
		_ = transaction.Rollback()
		_ = connection.Close()
		t.Fatal(err)
	}
	if err := transaction.Rollback(); err != nil {
		_ = connection.Close()
		t.Fatal(err)
	}
	_ = connection.Close()
	_, err = runLayoutRuntimeText(ctx, database, "layout_a", getSQL,
		rolledBackLayoutRevision)
	assertLayoutSQLState(t, err, "P0002")
}

func seedLayoutAdoptedFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE TABLE layout_a.legacy_system_resource_layout_source (
    id uuid PRIMARY KEY, payload jsonb NOT NULL
);
CREATE TABLE layout_b.legacy_system_resource_layout_source (
    id uuid PRIMARY KEY, payload jsonb NOT NULL
);
INSERT INTO layout_a.legacy_system_resource_layout_source
VALUES ('ffffffff-ffff-4fff-8fff-ffffffffffff','{"layout":"v1"}');
INSERT INTO layout_b.legacy_system_resource_layout_source
VALUES ('ffffffff-ffff-4fff-8fff-ffffffffffff','{"layout":"v1"}');`)
	if err != nil {
		t.Fatalf("seed adopted layout fixture: %v", err)
	}
}

func assertLayoutAdoptedFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	queries := []string{
		"SELECT pg_catalog.count(*) FROM " +
			"layout_a.legacy_system_resource_layout_source",
		"SELECT pg_catalog.count(*) FROM " +
			"layout_b.legacy_system_resource_layout_source",
	}
	for index, query := range queries {
		var count int
		err := database.QueryRowContext(ctx, query).Scan(&count)
		if err != nil || count != 1 {
			t.Fatalf("adopted layout source %d = %d: %v", index, count, err)
		}
	}
}
