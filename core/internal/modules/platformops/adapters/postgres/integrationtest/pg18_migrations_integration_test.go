//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops"
)

func TestPlatformOpsMigrationsAndRecoveryLifecyclePG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startPlatformOpsDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetPlatformOpsSchema(t, ctx, database)
			applyPlatformOpsOrigin(t, ctx, database, origin)
			setupRecoveryIdentity(t, ctx, database)
			exerciseRecoveryClaimShape(t, ctx, database)
			exerciseRecoveryProtection(t, ctx, database)
			exerciseRecoveryLifecycle(t, ctx, database)
			exerciseRecoveryHTTPArtifacts(t, ctx, database)
		})
	}
}

func resetPlatformOpsSchema(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	if _, err := database.ExecContext(ctx,
		`DROP SCHEMA IF EXISTS platform CASCADE; CREATE SCHEMA platform`); err != nil {
		t.Fatal(err)
	}
}

func applyPlatformOpsOrigin(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents, err := platformops.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	needle := "/migrations/platform/" + origin + "/"
	for _, document := range documents {
		if !strings.Contains(document.Path(), needle) {
			continue
		}
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
	}
}
