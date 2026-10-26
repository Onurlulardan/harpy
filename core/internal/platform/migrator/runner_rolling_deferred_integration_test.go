//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

const unactivatedReleaseID = "018f0000-0000-7000-8000-000000000601"
const unactivatedRunID = "018f0000-0000-7000-8000-000000000602"

func assertUnactivatedReleaseRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	active *Runner,
) {
	t.Helper()
	identity := readRollingIdentity(t, ctx, database)
	receipts := rollingTransactionReceipts(identity, unactivatedReleaseID, []string{
		"018f0000-0000-7000-8000-000000000603",
		"018f0000-0000-7000-8000-000000000604",
		"018f0000-0000-7000-8000-000000000605",
	}, "unactivated")
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	persistRollingRunEvidence(
		t, ctx, tx, active, identity, unactivatedReleaseID, unactivatedRunID, receipts,
	)
	record, err := active.currentReleaseGeneration(unactivatedRunID)
	if err != nil {
		t.Fatal(err)
	}
	record.ReleaseID = unactivatedReleaseID
	record.RuntimeGeneration = 5
	record.BootstrapReceiptRunID = receipts[0].RunID
	record.ExtensionReceiptRunID = receipts[1].RunID
	record.CleanupReceiptRunID = receipts[2].RunID
	if err := postgresstore.CreateReleaseGeneration(ctx, tx, record); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err == nil ||
		!strings.Contains(err.Error(), "not the finalized active contract") {
		t.Fatalf("unactivated release commit error = %v", err)
	}
	if _, err := postgresstore.FindReleaseGenerationByRelease(
		ctx, database, identity.EnvironmentID, unactivatedReleaseID,
	); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unactivated release persisted: %v", err)
	}
}
