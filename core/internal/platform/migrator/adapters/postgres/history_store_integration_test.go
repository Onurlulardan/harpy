//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	goosedb "github.com/pressly/goose/v3/database"
)

func TestHistoryStoreAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startHistoryIntegrationDatabase(t, ctx)
	selected := validSelectedMigrations()
	store, err := NewStore(testRunID, selected)
	if err != nil {
		t.Fatalf("create validated history store: %v", err)
	}
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin history creation transaction: %v", err)
	}
	if _, err := transaction.ExecContext(ctx, "SET LOCAL ROLE harpy_owner"); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("assume canonical migration owner role: %v", err)
	}
	if err := store.CreateVersionTable(ctx, transaction); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("create history table: %v", err)
	}
	anchor := goosedb.InsertRequest{Version: 0}
	if err := store.Insert(ctx, transaction, anchor); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("insert history anchor: %v", err)
	}
	if err := store.Insert(ctx, transaction, anchor); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("reinsert idempotent history anchor: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit history creation: %v", err)
	}
	assertHistoryTableAndAnchor(t, ctx, database, store)
	for _, migration := range selected {
		request := goosedb.InsertRequest{Version: migration.Version}
		if err := store.Insert(ctx, database, request); err != nil {
			t.Fatalf("insert migration %d metadata: %v", migration.Version, err)
		}
	}
	assertStoredMigrationMetadata(t, ctx, database, selected)
	assertHistoryReadSurface(t, ctx, database, store, selected)
	assertHistoryWriteDenials(t, ctx, database, store, selected)
}
