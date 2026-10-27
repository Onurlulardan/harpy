//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type smtpSettingsTrack struct {
	schema string
}

func prepareSMTPSettingsDatabase(t *testing.T, ctx context.Context,
	connection *pgx.Conn,
) []smtpSettingsTrack {
	t.Helper()
	_, err := connection.Exec(ctx, `CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_adopter NOLOGIN;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;`)
	if err != nil {
		t.Fatal(err)
	}
	applySMTPSettingsMigrations(t, ctx, connection, "platform", "")
	tracks := []smtpSettingsTrack{{"notifications_smtp_a"}, {"notifications_smtp_b"}}
	for _, track := range tracks {
		_, err = connection.Exec(ctx, "CREATE SCHEMA "+track.schema+
			" AUTHORIZATION harpy_owner; GRANT USAGE ON SCHEMA "+track.schema+
			" TO harpy_runtime, harpy_adopter;")
		if err != nil {
			t.Fatal(err)
		}
		applySMTPSettingsMigrations(t, ctx, connection, "tenant", track.schema)
		seedSMTPSettingsProvider(t, ctx, connection, track.schema)
	}
	return tracks
}

func applySMTPSettingsMigrations(t *testing.T, ctx context.Context,
	connection *pgx.Conn, scope, schema string,
) {
	t.Helper()
	directory := filepath.Join("..", "..", "migrations", scope, "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	applied := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".fresh.sql") {
			continue
		}
		document, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil || len(document) == 0 {
			t.Fatalf("fresh migration %s did not load: %v", entry.Name(), readErr)
		}
		applied++
		tx, beginErr := connection.Begin(ctx)
		if beginErr == nil && schema != "" {
			_, beginErr = tx.Exec(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)", schema+", pg_catalog")
		}
		if beginErr == nil {
			_, beginErr = tx.Exec(ctx, string(document))
		}
		if beginErr != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("apply %s: %v", entry.Name(), beginErr)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if applied == 0 {
		t.Fatalf("%s fresh migrations were not applied", scope)
	}
}

func beginSMTPSettingsTenant(t *testing.T, ctx context.Context,
	connection *pgx.Conn, schema string, runtime bool,
) pgx.Tx {
	t.Helper()
	tx, err := connection.Begin(ctx)
	if err == nil {
		_, err = tx.Exec(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
			schema+", pg_catalog")
	}
	if err == nil && runtime {
		_, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime")
	}
	if err != nil {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
		t.Fatal(err)
	}
	return tx
}

func smtpSettingsStore(tx pgx.Tx, tenant uuid.UUID) *Store {
	return &Store{command: true, tenantID: tenant,
		queryRow: func(ctx context.Context, sql string, arguments ...any) row {
			return tx.QueryRow(ctx, sql, arguments...)
		},
		query: func(ctx context.Context, sql string, arguments ...any) (rows, error) {
			return tx.Query(ctx, sql, arguments...)
		},
		execute: func(ctx context.Context, sql string, arguments ...any) (commandResult, error) {
			return tx.Exec(ctx, sql, arguments...)
		},
		notFound: func(err error) bool { return errors.Is(err, pgx.ErrNoRows) },
		allowed:  func(context.Context) bool { return true }}
}

func smtpDomainID(t *testing.T, value string) domain.ID {
	t.Helper()
	id, err := domain.ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
