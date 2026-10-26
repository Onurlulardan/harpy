//go:build integration

package migrator

import (
	"context"
	"testing"
	"time"
)

func TestProductionRoleFixtureMatchesExactPostgreSQL18Catalog(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	database := startBuiltEntrypointDatabase(t, ctx)
	assertProductionRoleFixture(t, ctx, database.admin)
}
