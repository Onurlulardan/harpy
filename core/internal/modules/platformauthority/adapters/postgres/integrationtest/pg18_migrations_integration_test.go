//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPlatformAuthorityMigrationsAndCASPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startPlatformAuthorityDatabase(t, ctx)
	resetPlatformSchema(t, ctx, database)
	applyAuthorityFreshDocuments(t, ctx, database)
	exerciseAuthorityCAS(t, ctx, database)
	exerciseDeliveryQueue(t, ctx, database)
}

func resetPlatformSchema(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	if _, err := database.ExecContext(ctx,
		`DROP SCHEMA IF EXISTS platform CASCADE;
CREATE SCHEMA platform;
GRANT USAGE ON SCHEMA platform TO harpy_runtime`); err != nil {
		t.Fatal(err)
	}
}

func applyAuthorityFreshDocuments(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	documents, err := platformauthority.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	applied := 0
	for _, document := range documents {
		if !strings.Contains(document.Path(), "/fresh/") {
			t.Fatalf("migration path = %s", document.Path())
		}
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			var pgError *pgconn.PgError
			if errors.As(err, &pgError) {
				t.Fatalf("apply %s: code=%s position=%d where=%q message=%s",
					document.Path(), pgError.Code, pgError.Position,
					pgError.Where, pgError.Message)
			}
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
		applied++
	}
	if applied == 0 {
		t.Fatal("no fresh platformauthority migrations")
	}
}
