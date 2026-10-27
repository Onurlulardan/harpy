//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func useBuiltQuietCompletionBatch(
	t *testing.T, signer builtCutoverSigner, adoption builtCutoverLeaseArtifacts,
	directory string, now time.Time,
) builtCutoverLeaseArtifacts {
	t.Helper()
	batch := buildBuiltCutoverBatch(t, adoption.identity)
	batch.Payload.Changes = []cutoverv1.Change{}
	payload, err := attestation.MarshalCanonical(batch.Payload)
	if err != nil {
		t.Fatal(err)
	}
	batch.BatchSHA256, err = attestation.DigestHex(payload)
	if err != nil {
		t.Fatal(err)
	}
	artifact := writeBuiltCutoverArtifact(t, directory, "quiet-batch.json", batch)
	verifiedLease, err := attestation.VerifyCutoverCredentialLease(
		adoption.leaseArtifact.document, adoption.leaseSignature.document, signer.root, now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := attestation.VerifyCutoverChangeBatch(artifact.document, verifiedLease); err != nil {
		t.Fatalf("verify quiet completion batch: %v", err)
	}
	adoption.batchArtifact = artifact
	return adoption
}

func removeBuiltJournalFixture(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts,
) {
	t.Helper()
	transaction, err := database.admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if _, err = transaction.Exec(ctx, "SET LOCAL session_replication_role='replica'"); err != nil {
		t.Fatal(err)
	}
	_, err = transaction.Exec(ctx,
		`DELETE FROM platform.cutover_change_journal WHERE operation_id=$1 AND run_id=$2`,
		adoption.identity.operationID, adoption.identity.runID)
	if err != nil {
		t.Fatalf("remove synthetic journal fixture: %v", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func buildBuiltQuietTerminalEvidence(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts, directory string, batch cutoverv1.ChangeBatch,
	batchRoot string,
) builtCutoverArtifact {
	t.Helper()
	coverageRoot, appliedCount, err := cutoverv1.TerminalJournalCoverageRootSHA256(nil)
	if err != nil || appliedCount != 0 {
		t.Fatalf("build empty journal root: count=%d err=%v", appliedCount, err)
	}
	finalLSN := *batch.Payload.Position.FinalLSN
	initial := cutoverv1.TerminalInitialCopyEvidence{
		ManifestSHA256: adoption.leaseArtifact.digest,
		SnapshotName:   batch.Payload.Position.SnapshotName,
		SnapshotLSN:    batch.Payload.Position.SnapshotLSN,
		RelationCount:  1, PartitionCount: 1, RowCount: 1, RootSHA256: digestMarker('d'),
	}
	checkpointCount := 1
	journalCount := int64(0)
	evidence := cutoverv1.TerminalEvidence{
		ContractVersion: "cutover-terminal-evidence.v1",
		OperationID:     adoption.identity.operationID, RunID: adoption.identity.runID,
		LeaseManifestSHA256:     adoption.leaseArtifact.digest,
		PlanSHA256:              adoption.lease.PlanSHA256,
		TransformManifestSHA256: adoption.lease.TransformManifestSHA256,
		TerminalState:           "completed", FinalLSN: &finalLSN, InitialCopy: &initial,
		CheckpointCount: &checkpointCount,
		Checkpoints: []cutoverv1.TerminalCheckpointEvidence{{
			RelationOrdinal: 1, CheckpointKey: batch.Payload.CheckpointKey,
			AppliedLSN: finalLSN, AppliedRows: 0, BatchCount: 1,
			BatchRootSHA256: batchRoot,
		}},
		JournalTransactionCount: &journalCount, AppliedChangeCount: &appliedCount,
		JournalCoverageRootSHA256: &coverageRoot,
	}
	assertBuiltQuietDatabaseState(t, ctx, database, adoption)
	if err := cutoverv1.ValidateTerminalEvidence(evidence); err != nil {
		t.Fatalf("validate quiet terminal evidence: %v", err)
	}
	return writeBuiltCutoverArtifact(t, directory, "terminal-evidence.json", evidence)
}

func assertBuiltQuietDatabaseState(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts,
) {
	t.Helper()
	var journals, changes, batches int
	err := database.admin.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM platform.cutover_change_journal WHERE operation_id=$1 AND run_id=$2),
 (SELECT count(*) FROM platform.applied_change_receipts WHERE operation_id=$1 AND run_id=$2),
 (SELECT count(*) FROM platform.cutover_batch_receipts WHERE operation_id=$1 AND run_id=$2 AND applied_count=0)`,
		adoption.identity.operationID, adoption.identity.runID).Scan(&journals, &changes, &batches)
	if err != nil || journals != 0 || changes != 0 || batches != 1 {
		t.Fatalf("quiet database evidence = %d %d %d, err=%v", journals, changes, batches, err)
	}
}
