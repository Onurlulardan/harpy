//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const layoutReferencedFieldID = "33333333-3333-4333-8333-333333333333"

func assertLayoutReferenceEvidence(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	assertLayoutReferenceCount(t, ctx, database, "layout_a", 1)
	assertLayoutReferenceCount(t, ctx, database, "layout_b", 0)
	assertLayoutRetiredRoutinesRevoked(t, ctx, database)
}

func assertLayoutReferenceCount(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	want int64,
) {
	t.Helper()
	connection, transaction, err := beginLayoutRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var count int64
	var digest []byte
	var owners int32
	err = transaction.QueryRowContext(
		ctx, referenceEvidenceSQL, "user", layoutReferencedFieldID,
	).Scan(&count, &digest, &owners)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if count != want || len(digest) != 32 || owners != 1 {
		_ = transaction.Rollback()
		t.Fatalf("reference evidence = (%d, %d, %d), want (%d, 32, 1)",
			count, len(digest), owners, want)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertLayoutRetiredRoutinesRevoked(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, version := range []string{"v1", "v2"} {
		statement := `SELECT system_resource_layout_publish_` + version + `(
            (pg_catalog.current_schemas(false))[2]::name,
            $1::uuid, $2::bigint, $3::text, $4::timestamptz
        )`
		_, err := runLayoutRuntimeText(
			ctx, database, "layout_a", statement,
			layoutSecondRevision, int64(2), layoutCatalog, layoutNow(9),
		)
		assertLayoutSQLState(t, err, "42501")
	}
	_, err := runLayoutRuntimeText(
		ctx, database, "layout_a",
		"SELECT system_resource_layout_resolve_v1($1::name, $2::text)",
		"layout_a", "user",
	)
	assertLayoutSQLState(t, err, "42501")
}
