//go:build integration

package v1adoption

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareNotificationDatabase(t *testing.T, ctx context.Context, connection *pgx.Conn) {
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
		t.Fatalf("apply platform notification migration: %v", err)
	}
}

func prepareNotificationTenant(t *testing.T, ctx context.Context, connection *pgx.Conn,
	schema string, source bool,
) {
	t.Helper()
	if _, err := connection.Exec(ctx, "CREATE SCHEMA "+schema+" AUTHORIZATION harpy_owner"); err != nil {
		t.Fatal(err)
	}
	if source {
		createV1NotificationSource(t, ctx, connection, schema)
	}
	applyNotificationMigrations(t, ctx, connection, schema)
	_, err := connection.Exec(ctx, `GRANT USAGE ON SCHEMA `+schema+` TO harpy_adopter;
GRANT SELECT ON ALL TABLES IN SCHEMA `+schema+` TO harpy_adopter;`)
	if err != nil {
		t.Fatal(err)
	}
}

func createV1NotificationSource(t *testing.T, ctx context.Context,
	connection *pgx.Conn, schema string,
) {
	t.Helper()
	_, err := connection.Exec(ctx, `CREATE TABLE `+schema+`.notifications (
 id uuid PRIMARY KEY, tenant_id uuid NOT NULL, user_id uuid, type varchar(50) NOT NULL,
 level varchar(20) NOT NULL DEFAULT 'info', title varchar(200) NOT NULL,
 message text NOT NULL, action varchar(100), entity_id uuid, entity varchar(100),
 data jsonb, channel varchar(20) NOT NULL DEFAULT 'email', is_read boolean DEFAULT false,
 created_at timestamp, read_at timestamp
);`)
	if err != nil {
		t.Fatal(err)
	}
}

func applyNotificationMigrations(t *testing.T, ctx context.Context,
	connection *pgx.Conn, schema string,
) {
	t.Helper()
	directory := filepath.Join("..", "..", "..", "migrations", "tenant", "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	applied := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".fresh.sql") {
			continue
		}
		document, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil || len(document) == 0 {
			t.Fatalf("fresh migration %s did not load: %v", entry.Name(), err)
		}
		applied++
		transaction, err := connection.Begin(ctx)
		if err == nil {
			_, err = transaction.Exec(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)", schema+", pg_catalog")
		}
		if err == nil {
			_, err = transaction.Exec(ctx, string(document))
		}
		if err != nil {
			_ = transaction.Rollback(ctx)
			t.Fatalf("apply %s: %v", entry.Name(), err)
		}
		if err = transaction.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if applied == 0 {
		t.Fatal("tenant fresh migrations were not applied")
	}
}
