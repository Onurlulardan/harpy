//go:build integration

package v1cutover

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareCutoverDatabase(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	_, err := connection.Exec(ctx, `CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_adopter NOLOGIN;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE TABLE platform.jobs (
 id uuid PRIMARY KEY, tenant_id uuid, authority_domain text NOT NULL,
 writer_origin text NOT NULL, state text NOT NULL
);`)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("..", "..", "..", "migrations", "platform", "fresh",
		"20260718220200_create_notification_projections_v1.fresh.sql")
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = connection.Exec(ctx, string(document)); err != nil {
		t.Fatal(err)
	}
	prepareCutoverTenant(t, ctx, connection, "cutover_a")
	prepareCutoverTenant(t, ctx, connection, "cutover_b")
}

func prepareCutoverTenant(t *testing.T, ctx context.Context, connection *pgx.Conn,
	schema string,
) {
	t.Helper()
	if _, err := connection.Exec(ctx, "CREATE SCHEMA "+schema+" AUTHORIZATION harpy_owner"); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join("..", "..", "..", "migrations", "tenant", "v1_adopted")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".v1_adopted.sql") {
			continue
		}
		document, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		transaction, beginErr := connection.Begin(ctx)
		if beginErr == nil {
			_, beginErr = transaction.Exec(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)", schema+", pg_catalog")
		}
		if beginErr == nil {
			_, beginErr = transaction.Exec(ctx, string(document))
		}
		if beginErr != nil {
			_ = transaction.Rollback(ctx)
			t.Fatalf("apply %s: %v", entry.Name(), beginErr)
		}
		if beginErr = transaction.Commit(ctx); beginErr != nil {
			t.Fatal(beginErr)
		}
	}
}
