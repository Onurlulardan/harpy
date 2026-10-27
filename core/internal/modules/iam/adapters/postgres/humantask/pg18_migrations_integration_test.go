//go:build integration

package humantask

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type iamHumanMigration struct {
	name string
	body []byte
}

func loadIAMHumanMigrations(t *testing.T, origin string) []iamHumanMigration {
	t.Helper()
	directory := filepath.Join("..", "..", "..", "migrations", "tenant", origin)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	suffix := "." + origin + ".sql"
	result := make([]iamHumanMigration, 0, 5)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) ||
			!strings.HasPrefix(entry.Name(), "2026071805050") {
			continue
		}
		body, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		result = append(result, iamHumanMigration{name: entry.Name(), body: body})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].name < result[right].name })
	if len(result) != 5 {
		t.Fatalf("IAM human migration count = %d", len(result))
	}
	return result
}

func applyIAMHumanMigrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	for _, migration := range loadIAMHumanMigrations(t, origin) {
		for _, schema := range []string{"iam_human_a", "iam_human_b"} {
			transaction, err := database.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = transaction.ExecContext(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)", schema+", pg_catalog")
			if err == nil {
				_, err = transaction.ExecContext(ctx, string(migration.body))
			}
			if err != nil {
				_ = transaction.Rollback()
				t.Fatalf("apply %s/%s: %v", schema, migration.name, err)
			}
			if err = transaction.Commit(); err != nil {
				t.Fatal(err)
			}
		}
	}
}
