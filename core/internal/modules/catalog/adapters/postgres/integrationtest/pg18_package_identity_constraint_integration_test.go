//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func assertCatalogPackageIdentityConstraints(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	assertPackageConstraintSet(t, ctx, database, "platform", 3)
	assertPackageConstraintSet(t, ctx, database, "catalog_a", 12)
	at := time.Date(2026, 7, 31, 11, 0, 0, 123000, time.UTC)
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO catalog_a.catalog_installations(
installation_id,package_id,active_version,state,reference_epoch,row_version,updated_at)
VALUES('019fa200-0000-7000-8000-000000000001',
'e3000000-0000-0000-0000-000000000001',NULL,'INSTALLING',1,1,$1)`, at)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("opaque package identity rejected: %v", err)
	}
	if err = transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `INSERT INTO catalog_a.catalog_installations(
installation_id,package_id,active_version,state,reference_epoch,row_version,updated_at)
VALUES('019fa200-0000-7000-8000-000000000002',
'00000000-0000-0000-0000-000000000000',NULL,'INSTALLING',1,1,$1)`, at)
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) ||
		pgError.ConstraintName != "catalog_installations_package_id_nonzero" {
		t.Fatalf("nil package identity error = %v", err)
	}
}

func assertPackageConstraintSet(
	t *testing.T, ctx context.Context, database *sql.DB, schema string, want int,
) {
	t.Helper()
	var count int
	var validated bool
	var definitions string
	err := database.QueryRowContext(ctx, `SELECT count(*),bool_and(c.convalidated),
string_agg(pg_catalog.pg_get_constraintdef(c.oid),'|')
FROM pg_catalog.pg_constraint AS c
JOIN pg_catalog.pg_class AS relation ON relation.oid=c.conrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
WHERE namespace.nspname=$1 AND c.conname LIKE 'catalog_%package_id_nonzero'`,
		schema).Scan(&count, &validated, &definitions)
	if err != nil || count != want || !validated ||
		strings.Contains(definitions, "uuid_extract_version") {
		t.Fatalf("%s package constraints=%d/%v %s err=%v",
			schema, count, validated, definitions, err)
	}
}
