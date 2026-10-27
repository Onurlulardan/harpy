//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestDocumentsMigrationsAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	database := startDocumentsPG18(t, ctx)
	prepareDocumentsPG18(t, ctx, database)
	applyDocumentsMigrations(t, ctx, database)
	for _, schema := range documentsSchemas {
		t.Run(schema, func(t *testing.T) {
			assertDocumentsSchema(t, ctx, database, schema)
		})
	}
	assertDocumentsTenantIsolation(t, ctx, database)
	assertDocumentsAdoptionSQL(t, ctx, database)
	assertDocumentsUploadSQL(t, ctx, database)
	assertDocumentsArchiveSQL(t, ctx, database)
	assertDocumentsReferenceSQL(t, ctx, database)
	assertDocumentsPurgeSQL(t, ctx, database)
	for _, schema := range []string{"documents_fresh_a", "documents_adopted_a"} {
		assertDocumentsFormRecordLinks(t, ctx, database, schema)
		assertDocumentsFormsCleanupSQL(t, ctx, database, schema)
		assertDocumentsKnowledgeFileSQL(t, ctx, database, schema)
	}
}

func assertDocumentsSchema(t *testing.T, ctx context.Context, database *sql.DB,
	schema string,
) {
	t.Helper()
	var tables, functions int
	err := database.QueryRowContext(ctx, `SELECT
  (SELECT count(*) FROM pg_catalog.pg_tables WHERE schemaname = $1),
  (SELECT count(*) FROM pg_catalog.pg_proc p
   JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
   WHERE n.nspname = $1)`, schema).Scan(&tables, &functions)
	if err != nil {
		t.Fatal(err)
	}
	if tables != 30 || functions != 50 {
		t.Fatalf("%s tables/functions = %d/%d, want 30/50", schema, tables, functions)
	}
	insertDocumentsFolder(t, ctx, database, schema,
		"018f0000-0000-7000-8000-000000000101")
}

func insertDocumentsFolder(t *testing.T, ctx context.Context, database *sql.DB,
	schema, id string,
) {
	t.Helper()
	statement := `INSERT INTO ` + schema + `.documents_folders
 (id, parent_id, owner_id, display_name, lifecycle_state, row_version,
  created_at, updated_at, trashed_at)
 VALUES ($1, NULL, '22222222-2222-4222-8222-222222222222',
  'Evidence', 'ACTIVE', 1, $2, $2, NULL)`
	if _, err := database.ExecContext(ctx, statement, id, time.Now().UTC()); err != nil {
		t.Fatalf("insert folder into %s: %v", schema, err)
	}
}

func assertDocumentsTenantIsolation(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var count int
	err := database.QueryRowContext(ctx,
		"SELECT count(*) FROM documents_fresh_b.documents_folders").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("fresh tenant B folder count = %d, want own fixture only", count)
	}
}
