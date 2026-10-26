//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func buildBuiltTerminalEvidence(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts, directory string,
) builtCutoverArtifact {
	t.Helper()
	if adoption.identity.terminalState != "completed" {
		return buildBuiltAbortTerminalEvidence(t, adoption, directory)
	}
	var batch cutoverv1.ChangeBatch
	if err := attestation.DecodeStrict(adoption.batchArtifact.document, &batch); err != nil {
		t.Fatal(err)
	}
	leaf := readBuiltTerminalBatchLeaf(t, ctx, database, adoption, batch)
	batchRoot, err := cutoverv1.TerminalBatchRootSHA256([]cutoverv1.TerminalBatchLeaf{leaf})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Payload.Changes) == 0 {
		return buildBuiltQuietTerminalEvidence(
			t, ctx, database, adoption, directory, batch, batchRoot,
		)
	}
	coverage := builtTerminalCoverage(batch)
	coverageRoot, appliedCount, err := cutoverv1.TerminalJournalCoverageRootSHA256(
		[]cutoverv1.TerminalJournalCoverageLeaf{coverage},
	)
	if err != nil {
		t.Fatal(err)
	}
	finalLSN := *batch.Payload.Position.FinalLSN
	initial := cutoverv1.TerminalInitialCopyEvidence{
		ManifestSHA256: adoption.leaseArtifact.digest,
		SnapshotName:   batch.Payload.Position.SnapshotName,
		SnapshotLSN:    batch.Payload.Position.SnapshotLSN,
		RelationCount:  1, PartitionCount: 1, RowCount: 1,
		RootSHA256: digestMarker('d'),
	}
	checkpointCount, journalCount := 1, int64(1)
	evidence := cutoverv1.TerminalEvidence{
		ContractVersion: "cutover-terminal-evidence.v1",
		OperationID:     adoption.identity.operationID, RunID: adoption.identity.runID,
		LeaseManifestSHA256:     adoption.leaseArtifact.digest,
		PlanSHA256:              adoption.lease.PlanSHA256,
		TransformManifestSHA256: adoption.lease.TransformManifestSHA256,
		TerminalState:           "completed", FinalLSN: &finalLSN,
		InitialCopy: &initial, CheckpointCount: &checkpointCount,
		Checkpoints: []cutoverv1.TerminalCheckpointEvidence{{
			RelationOrdinal: batch.Payload.RelationOrdinal,
			CheckpointKey:   batch.Payload.CheckpointKey,
			AppliedLSN:      batch.Payload.Position.CommitLSN,
			AppliedRows:     int64(len(batch.Payload.Changes)), BatchCount: 1,
			BatchRootSHA256: batchRoot,
		}},
		JournalTransactionCount: &journalCount, AppliedChangeCount: &appliedCount,
		JournalCoverageRootSHA256: &coverageRoot,
	}
	assertBuiltTerminalDatabaseMatches(t, ctx, database, adoption, evidence, coverage)
	if err := cutoverv1.ValidateTerminalEvidence(evidence); err != nil {
		t.Fatalf("validate built terminal evidence: %v", err)
	}
	artifact := writeBuiltCutoverArtifact(t, directory, "terminal-evidence.json", evidence)
	return artifact
}

func buildBuiltAbortTerminalEvidence(
	t *testing.T, adoption builtCutoverLeaseArtifacts, directory string,
) builtCutoverArtifact {
	t.Helper()
	code := "verification_failed"
	if adoption.identity.terminalState == "cancelled" {
		code = "operator_cancelled"
	}
	outcomeDigest := digestMarker('e')
	evidence := cutoverv1.TerminalEvidence{
		ContractVersion: "cutover-terminal-evidence.v1",
		OperationID:     adoption.identity.operationID, RunID: adoption.identity.runID,
		LeaseManifestSHA256:     adoption.leaseArtifact.digest,
		PlanSHA256:              adoption.lease.PlanSHA256,
		TransformManifestSHA256: adoption.lease.TransformManifestSHA256,
		TerminalState:           adoption.identity.terminalState,
		OutcomeCode:             &code, OutcomeEvidenceSHA256: &outcomeDigest,
	}
	if err := cutoverv1.ValidateTerminalEvidence(evidence); err != nil {
		t.Fatalf("validate built abort terminal evidence: %v", err)
	}
	return writeBuiltCutoverArtifact(t, directory, "terminal-evidence.json", evidence)
}

func readBuiltTerminalBatchLeaf(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts, batch cutoverv1.ChangeBatch,
) cutoverv1.TerminalBatchLeaf {
	t.Helper()
	var leaf cutoverv1.TerminalBatchLeaf
	err := database.admin.QueryRow(ctx, `SELECT batch_ordinal,source_xid::text,
source_commit_lsn::text,batch_sha256,applied_count,key_state_sha256,state_sha256
FROM platform.cutover_batch_receipts
WHERE operation_id=$1 AND run_id=$2 AND relation_ordinal=$3 AND checkpoint_key=$4`,
		adoption.identity.operationID, adoption.identity.runID,
		batch.Payload.RelationOrdinal, batch.Payload.CheckpointKey,
	).Scan(&leaf.BatchOrdinal, &leaf.SourceXID, &leaf.SourceCommitLSN,
		&leaf.BatchSHA256, &leaf.AppliedCount, &leaf.KeyStateSHA256, &leaf.StateSHA256)
	if err != nil {
		t.Fatalf("read built terminal batch proof: %v", err)
	}
	position := batch.Payload.Position
	if leaf.BatchOrdinal != position.BatchOrdinal || leaf.SourceXID != position.SourceXID ||
		leaf.SourceCommitLSN != position.CommitLSN || leaf.BatchSHA256 != batch.BatchSHA256 ||
		leaf.AppliedCount != int64(len(batch.Payload.Changes)) {
		t.Fatal("built terminal batch proof differs from decoder evidence")
	}
	return leaf
}

func builtTerminalCoverage(batch cutoverv1.ChangeBatch) cutoverv1.TerminalJournalCoverageLeaf {
	changes := make([]cutoverv1.TerminalChangeLeaf, len(batch.Payload.Changes))
	for index, change := range batch.Payload.Changes {
		mutation := sha256.Sum256([]byte(batch.BatchSHA256 + ":" + strconv.Itoa(change.Ordinal)))
		changes[index] = cutoverv1.TerminalChangeLeaf{
			RelationOrdinal: batch.Payload.RelationOrdinal, ChangeOrdinal: change.Ordinal,
			CheckpointKey: batch.Payload.CheckpointKey, MutationKind: change.Kind,
			MutationSHA256: hex.EncodeToString(mutation[:]),
		}
	}
	return cutoverv1.TerminalJournalCoverageLeaf{
		SourceXID:       batch.Payload.Position.SourceXID,
		SourceCommitLSN: batch.Payload.Position.CommitLSN, Changes: changes,
	}
}

func assertBuiltTerminalDatabaseMatches(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts, evidence cutoverv1.TerminalEvidence,
	coverage cutoverv1.TerminalJournalCoverageLeaf,
) {
	t.Helper()
	var journalCount, appliedCount int64
	var mutationSHA string
	err := database.admin.QueryRow(ctx, `SELECT
 (SELECT count(DISTINCT source_xid) FROM platform.cutover_change_journal WHERE operation_id=$1 AND run_id=$2),
 (SELECT count(*) FROM platform.applied_change_receipts WHERE operation_id=$1 AND run_id=$2),
 (SELECT mutation_sha256 FROM platform.applied_change_receipts WHERE operation_id=$1 AND run_id=$2)`,
		adoption.identity.operationID, adoption.identity.runID,
	).Scan(&journalCount, &appliedCount, &mutationSHA)
	if err != nil || evidence.JournalTransactionCount == nil || evidence.AppliedChangeCount == nil ||
		journalCount != *evidence.JournalTransactionCount ||
		appliedCount != *evidence.AppliedChangeCount || mutationSHA != coverage.Changes[0].MutationSHA256 {
		t.Fatalf("built terminal journal proof differs: %v", err)
	}
}
