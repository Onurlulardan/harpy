//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type workMigrationDocument struct {
	name    string
	content []byte
}

func loadWorkMigrations(t *testing.T) []workMigrationDocument {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(workingDirectory, "..", "..", "..",
		"migrations", "tenant", "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	documents := make([]workMigrationDocument, 0, 32)
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".fresh.sql") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if len(content) == 0 {
			t.Fatalf("fresh migration empty: %s", entry.Name())
		}
		documents = append(documents, workMigrationDocument{entry.Name(), content})
	}
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].name < documents[right].name
	})
	if len(documents) != 32 {
		t.Fatalf("Work fresh migration count = %d, want 32", len(documents))
	}
	return documents
}

func prepareWorkSchemas(t *testing.T, ctx context.Context,
	database *sql.DB,
) map[string]legacyFixture {
	t.Helper()
	prepareWorkCutoverPlatform(t, ctx, database)
	for _, schema := range []string{"work_fresh_a", "work_fresh_b", "work_adopted_a", "work_adopted_b"} {
		if _, err := database.ExecContext(ctx, "CREATE SCHEMA "+schema+" AUTHORIZATION harpy_owner"); err != nil {
			t.Fatal(err)
		}
	}
	for _, schema := range []string{"work_adopted_a", "work_adopted_b"} {
		workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
			if _, err := transaction.ExecContext(ctx, workLegacyDDL); err != nil {
				t.Fatal(err)
			}
		})
	}
	fixtures := map[string]legacyFixture{
		"work_adopted_a": seedLegacyWork(t, ctx, database, "work_adopted_a", "Adopted A", 21),
		"work_adopted_b": seedLegacyWork(t, ctx, database, "work_adopted_b", "Adopted B", 22),
	}
	documents := loadWorkMigrations(t)
	for _, schema := range []string{"work_fresh_a", "work_fresh_b", "work_adopted_a", "work_adopted_b"} {
		prepareWorkTenantIdentity(t, ctx, database, schema, workIdentityOrigin(schema))
		for _, document := range documents {
			workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
				if _, err := transaction.ExecContext(ctx, string(document.content)); err != nil {
					t.Fatalf("apply %s to %s: %v", document.name, schema, err)
				}
			})
		}
	}
	return fixtures
}

func workIdentityOrigin(schema string) string {
	if schema == "work_adopted_a" || schema == "work_adopted_b" {
		return "v1_adopted"
	}
	return "new"
}

func workInSchema(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, run func(*sql.Tx),
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path', $1, true)", schema+",pg_catalog"); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	run(transaction)
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
