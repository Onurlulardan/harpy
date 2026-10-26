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
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetPlatformSchema(t, ctx, database)
			applyAuthorityOrigin(t, ctx, database, origin)
			exerciseAuthorityCAS(t, ctx, database)
			exerciseDeliveryQueue(t, ctx, database)
		})
	}
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

func applyAuthorityOrigin(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents, err := platformauthority.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	needle := "/migrations/platform/" + origin + "/"
	for _, document := range documents {
		if !strings.Contains(document.Path(), needle) {
			continue
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
	}
}
