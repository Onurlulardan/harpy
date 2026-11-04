//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func applyBillingCutoverPlatformMigrations(
	t *testing.T, _ context.Context, _ *sql.DB, _ string,
) {
	t.Helper()
	t.Fatal("billing cutover platform SQL is not a Goose source")
}
