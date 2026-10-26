//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func seedRetentionCycleRows(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	ids retentionCycleWorkIDs,
	inboxEvent uuid.UUID,
) {
	t.Helper()
	seedRetentionJobsAndReceipts(t, ctx, database, ids, inboxEvent)
	seedRetentionDeadLetters(t, ctx, database, ids)
}

func seedRetentionJobsAndReceipts(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	ids retentionCycleWorkIDs,
	inboxEvent uuid.UUID,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin retention row fixture: %v", err)
	}
	defer transaction.Rollback()
	if err = setRetentionSeedTriggers(ctx, transaction, false); err != nil {
		t.Fatalf("disable retention fixture triggers: %v", err)
	}
	insertRetentionJobRows(t, ctx, transaction, ids)
	insertRetentionReceiptRow(t, ctx, transaction, ids.purgeReceipt)
	ageRetentionInboxRows(t, ctx, transaction, inboxEvent)
	if err = setRetentionSeedTriggers(ctx, transaction, true); err != nil {
		t.Fatalf("restore retention fixture triggers: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit retention row fixture: %v", err)
	}
}

func setRetentionSeedTriggers(
	ctx context.Context, transaction *sql.Tx, enable bool,
) error {
	statements := []string{
		"ALTER TABLE platform.jobs DISABLE TRIGGER ALL",
		"ALTER TABLE platform.inbox DISABLE TRIGGER ALL",
		"ALTER TABLE platform.idempotency_receipts DISABLE TRIGGER ALL",
	}
	if enable {
		statements = []string{
			"ALTER TABLE platform.idempotency_receipts ENABLE TRIGGER ALL",
			"ALTER TABLE platform.inbox ENABLE TRIGGER ALL",
			"ALTER TABLE platform.jobs ENABLE TRIGGER ALL",
		}
	}
	for _, statement := range statements {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
