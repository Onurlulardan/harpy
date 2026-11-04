//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func applyAnalyticsCutoverDependencies(t *testing.T, _ context.Context, _ *sql.DB) {
	t.Helper()
	t.Fatal("analytics cutover dependency SQL is not a Goose source")
}
