//go:build integration

package migrator

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareBuiltEntrypointCluster(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	for _, statement := range productionRoleFixtureStatements(builtEntrypointDatabaseName) {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("prepare built entrypoint role contract: %v", err)
		}
	}
}
