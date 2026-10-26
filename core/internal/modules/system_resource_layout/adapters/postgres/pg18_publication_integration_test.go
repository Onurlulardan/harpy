//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertPublishedLayoutIsTerminal(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	definition string,
) {
	t.Helper()
	_, err := runLayoutRuntimeText(ctx, database, "layout_a", updateSQL,
		layoutFirstRevision, 3, definition, layoutDigest(definition), layoutNow(5))
	assertLayoutSQLState(t, err, "40001")
	_, err = runLayoutRuntimeText(ctx, database, "layout_a", deleteSQL,
		layoutFirstRevision, 3)
	assertLayoutSQLState(t, err, "40001")
}

func assertLayoutList(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	document, err := runLayoutRuntimeText(ctx, database, "layout_a", listSQL, "user")
	if err != nil {
		t.Fatal(err)
	}
	revisions, err := decodeRevisionList([]byte(document))
	if err != nil || len(revisions) != 2 || revisions[0].Version != 1 ||
		revisions[1].Version != 2 {
		t.Fatalf("listed layouts = %+v: %v", revisions, err)
	}
}
