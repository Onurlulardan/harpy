//go:build integration

package migrator

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareBuiltEntrypointExtensions(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	statements := []string{
		"CREATE SCHEMA platform_ext AUTHORIZATION harpy_owner",
		"SET ROLE harpy_extension_admin",
		"CREATE EXTENSION pg_trgm WITH SCHEMA platform_ext",
		"CREATE EXTENSION vector WITH SCHEMA platform_ext",
		"RESET ROLE",
	}
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("prepare built entrypoint extension contract: %v", err)
		}
	}
}
