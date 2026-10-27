//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"

	documentsmodule "github.com/Onurlulardan/harpy/core/internal/modules/documents"
)

var documentsSchemas = []string{
	"documents_fresh_a", "documents_fresh_b",
	"documents_adopted_a", "documents_adopted_b",
}

func prepareDocumentsPG18(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	statement := `CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_backup_reader NOLOGIN NOINHERIT;
CREATE ROLE harpy_backup LOGIN NOINHERIT;
GRANT harpy_backup_reader TO harpy_backup;
CREATE ROLE harpy_restore LOGIN NOINHERIT;
CREATE ROLE harpy_adopter NOLOGIN NOINHERIT CONNECTION LIMIT 0;
CREATE ROLE harpy_cutover_decoder NOLOGIN NOINHERIT CONNECTION LIMIT 0;
CREATE ROLE harpy_cutover_finalizer NOLOGIN NOINHERIT CONNECTION LIMIT 0;`
	for _, schema := range documentsSchemas {
		statement += "\nCREATE SCHEMA " + schema + ";"
	}
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("prepare documents database: %v", err)
	}
}

func applyDocumentsMigrations(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	documents, err := documentsmodule.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].Path() < documents[right].Path()
	})
	for _, schema := range documentsSchemas {
		origin := "fresh"
		if strings.Contains(schema, "adopted") {
			origin = "v1_adopted"
		}
		for _, document := range documents {
			if !strings.Contains(document.Path(), "/"+origin+"/") {
				continue
			}
			applyDocumentsMigration(t, ctx, database, schema,
				document.Path(), document.Content())
		}
	}
}

func applyDocumentsMigration(t *testing.T, ctx context.Context, database *sql.DB,
	schema, path string, body []byte,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	var bound string
	if err = transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.set_config('search_path', $1, true)",
		schema+", pg_catalog").Scan(&bound); err == nil {
		_, err = transaction.ExecContext(ctx, string(body))
	}
	if err != nil {
		t.Fatalf("apply %s to %s: %v", path, schema, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
