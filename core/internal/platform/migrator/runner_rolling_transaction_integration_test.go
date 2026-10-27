//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

const outsideReleaseID = "018f0000-0000-7000-8000-000000000501"
const outsideRunID = "018f0000-0000-7000-8000-000000000502"

func assertOutOfTransactionReleaseRejected(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	active *Runner,
) {
	t.Helper()
	identity := readRollingIdentity(t, ctx, database)
	receipts := rollingOutsideReceipts(identity)
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	persistRollingRunEvidence(
		t, ctx, tx, active, identity, outsideReleaseID, outsideRunID, receipts,
	)
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	record, err := active.currentReleaseGeneration(outsideRunID)
	if err != nil {
		t.Fatal(err)
	}
	record.ReleaseID = outsideReleaseID
	record.RuntimeGeneration = 4
	record.BootstrapReceiptRunID = receipts[0].RunID
	record.ExtensionReceiptRunID = receipts[1].RunID
	record.CleanupReceiptRunID = receipts[2].RunID
	err = postgresstore.CreateReleaseGeneration(ctx, database, record)
	if err == nil || !strings.Contains(err.Error(), "outside exact finalization") {
		t.Fatalf("out-of-transaction release generation error = %v", err)
	}
	if _, err := postgresstore.FindReleaseGenerationByRelease(
		ctx, database, identity.EnvironmentID, outsideReleaseID,
	); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("out-of-transaction release persisted: %v", err)
	}
}

func persistRollingRunEvidence(
	t *testing.T,
	ctx context.Context,
	tx *sql.Tx,
	active *Runner,
	identity postgresstore.EnvironmentIdentityRecord,
	releaseID string,
	runID string,
	receipts []postgresstore.ReceiptEvidenceRecord,
) {
	t.Helper()
	input := postgresstore.CreateMigrationRunInput{
		RunID: runID, ReleaseID: releaseID,
		EnvironmentID: identity.EnvironmentID, DatabaseOrigin: identity.DatabaseOrigin,
		Command: "up", MigrationManifestSHA256: active.identity.manifestSHA256,
		ExtensionManifestSHA256:  active.identity.extensionSHA256,
		RolesManifestSHA256:      integrationDigest("outside-roles"),
		PrivilegesManifestSHA256: integrationDigest("outside-privileges"),
		OwnershipManifestSHA256:  integrationDigest("outside-ownership"),
	}
	if err := postgresstore.CreateMigrationRun(ctx, tx, input); err != nil {
		t.Fatal(err)
	}
	if err := postgresstore.StartMigrationRun(ctx, tx, runID); err != nil {
		t.Fatal(err)
	}
	for _, receipt := range receipts {
		if err := postgresstore.CreateReceiptEvidence(ctx, tx, receipt); err != nil {
			t.Fatal(err)
		}
	}
	if err := postgresstore.CompleteMigrationRun(ctx, tx, runID); err != nil {
		t.Fatal(err)
	}
}

func readRollingIdentity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) postgresstore.EnvironmentIdentityRecord {
	t.Helper()
	record, err := postgresstore.GetEnvironmentIdentity(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func rollingOutsideReceipts(
	identity postgresstore.EnvironmentIdentityRecord,
) []postgresstore.ReceiptEvidenceRecord {
	return rollingTransactionReceipts(identity, outsideReleaseID, []string{
		"018f0000-0000-7000-8000-000000000503",
		"018f0000-0000-7000-8000-000000000504",
		"018f0000-0000-7000-8000-000000000505",
	}, "outside")
}

func rollingTransactionReceipts(
	identity postgresstore.EnvironmentIdentityRecord,
	releaseID string,
	runIDs []string,
	label string,
) []postgresstore.ReceiptEvidenceRecord {
	kinds := []string{"cluster_bootstrap", "extension_admin", "credential_cleanup"}
	records := make([]postgresstore.ReceiptEvidenceRecord, 0, len(kinds))
	for index, kind := range kinds {
		document := []byte(fmt.Sprintf("{\"%s_%c\":1}", label, 'a'+index))
		records = append(records, postgresstore.ReceiptEvidenceRecord{
			Kind: kind, RunID: runIDs[index], EnvironmentID: identity.EnvironmentID,
			ReleaseID: releaseID, DatabaseOrigin: identity.DatabaseOrigin,
			DatabaseSystemID: identity.DatabaseSystemID, DatabaseOID: identity.DatabaseOID,
			OperationDigest:   integrationDigest(label + "-" + kind),
			CanonicalDocument: document, CanonicalSHA256: integrationBytesDigest(document),
			SignatureKeyID: "release_key", IssuedAt: time.Now().UTC().Add(-time.Minute),
		})
	}
	return records
}
