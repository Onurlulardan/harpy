package attestation

import (
	"bytes"
	"strings"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func TestVerifyCutoverTerminalEvidenceBindsExactSignedPlan(t *testing.T) {
	root, signer, lease, leaseSHA := verifiedCutoverLeaseFixture(t)
	document, digest := terminalEvidenceDocument(t, leaseSHA)
	finalizer := terminalFinalizer(t, root, signer, lease, leaseSHA, digest, "completed")
	verified, err := VerifyCutoverTerminalEvidence(document, lease, finalizer)
	if err != nil {
		t.Fatalf("valid terminal evidence rejected: %v", err)
	}
	if verified.SHA256() != digest || !bytes.Equal(verified.CanonicalDocument(), document) {
		t.Fatal("verified terminal evidence lost canonical identity")
	}
	mutated := verified.Value()
	mutated.Checkpoints[0].CheckpointKey = "mutated"
	*mutated.FinalLSN = "0/99"
	mutated.InitialCopy.SnapshotName = "mutated"
	if verified.Value().Checkpoints[0].CheckpointKey != "main" ||
		*verified.Value().FinalLSN != "0/20" ||
		verified.Value().InitialCopy.SnapshotName != "snapshot-main" {
		t.Fatal("verified terminal evidence leaked mutable checkpoints")
	}
}

func TestVerifyCutoverTerminalEvidenceRejectsDrift(t *testing.T) {
	root, signer, lease, leaseSHA := verifiedCutoverLeaseFixture(t)
	document, digest := terminalEvidenceDocument(t, leaseSHA)
	finalizer := terminalFinalizer(t, root, signer, lease, leaseSHA, digest, "completed")
	tests := []struct {
		name           string
		document       []byte
		wrongFinalizer bool
	}{
		{"unverified lease", document, false},
		{"wrong digest", document, true},
		{"non canonical bytes", append(append([]byte(nil), document...), '\n'), false},
		{"unknown member", bytes.Replace(document, []byte(`{"applied_change_count"`),
			[]byte(`{"unknown":true,"applied_change_count"`), 1), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := lease
			if test.name == "unverified lease" {
				candidate = VerifiedCutoverCredentialLease{}
			}
			candidateFinalizer := finalizer
			if test.wrongFinalizer {
				candidateFinalizer = terminalFinalizer(t, root, signer, lease, leaseSHA,
					strings.Repeat("f", 64), "completed")
			}
			if _, err := VerifyCutoverTerminalEvidence(
				test.document, candidate, candidateFinalizer,
			); err == nil {
				t.Fatal("invalid terminal evidence accepted")
			}
		})
	}
}

func TestVerifyCutoverTerminalEvidenceRejectsMissingOrExtraExpectedKey(t *testing.T) {
	root, signer, lease, leaseSHA := verifiedCutoverLeaseFixture(t)
	for _, key := range []string{"unexpected", ""} {
		value := terminalEvidenceFixture(leaseSHA)
		if key == "" {
			value.Checkpoints = nil
			zero := 0
			value.CheckpointCount = &zero
		} else {
			value.Checkpoints[0].CheckpointKey = key
		}
		document, err := MarshalCanonical(value)
		if err != nil {
			t.Fatal(err)
		}
		digest, err := DigestHex(document)
		if err != nil {
			t.Fatal(err)
		}
		finalizer := terminalFinalizer(t, root, signer, lease, leaseSHA, digest, "completed")
		if _, err := VerifyCutoverTerminalEvidence(document, lease, finalizer); err == nil {
			t.Fatal("terminal evidence checkpoint drift accepted")
		}
	}
}

func TestVerifyCutoverTerminalEvidenceBindsAbortStateAndClonesOutcome(t *testing.T) {
	root, signer, lease, leaseSHA := verifiedCutoverLeaseFixture(t)
	digestValue := strings.Repeat("a", 64)
	code := "verification_failed"
	value := cutoverv1.TerminalEvidence{
		ContractVersion: "cutover-terminal-evidence.v1",
		OperationID:     attestationOperationID, RunID: attestationRunID,
		LeaseManifestSHA256: leaseSHA, PlanSHA256: digestValue,
		TransformManifestSHA256: digestValue, TerminalState: "failed",
		OutcomeCode: &code, OutcomeEvidenceSHA256: &digestValue,
	}
	document, err := MarshalCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	finalizer := terminalFinalizer(t, root, signer, lease, leaseSHA, digest, "failed")
	verified, err := VerifyCutoverTerminalEvidence(document, lease, finalizer)
	if err != nil {
		t.Fatal(err)
	}
	mutated := verified.Value()
	*mutated.OutcomeCode = "source_failed"
	if *verified.Value().OutcomeCode != "verification_failed" {
		t.Fatal("verified abort outcome leaked mutable state")
	}
	wrongState := terminalFinalizer(t, root, signer, lease, leaseSHA, digest, "cancelled")
	if _, err := VerifyCutoverTerminalEvidence(document, lease, wrongState); err == nil {
		t.Fatal("terminal evidence accepted a different finalizer state")
	}
}

func terminalFinalizer(
	t *testing.T, root TrustRoot, signer Signer, adoption VerifiedCutoverCredentialLease,
	leaseSHA, terminalDigest, terminalState string,
) VerifiedCutoverFinalizerLease {
	t.Helper()
	seal, err := VerifyCutoverSessionSeal(
		signCutoverSeal(t, signer, attestationSessionSeal(leaseSHA)),
		adoption, root, cutoverTestNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	cleanup, err := VerifyCutoverCredentialCleanup(
		signCutoverCleanup(t, signer, attestationCleanup(leaseSHA, seal.ReceiptSHA256())),
		adoption, seal, root, cutoverTestNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	return verifiedFinalizerFixtureForTerminal(
		t, signer, root, adoption, cleanup, terminalDigest, terminalState,
	)
}

func terminalEvidenceDocument(t *testing.T, leaseSHA string) ([]byte, string) {
	t.Helper()
	document, err := MarshalCanonical(terminalEvidenceFixture(leaseSHA))
	if err != nil {
		t.Fatal(err)
	}
	digest, err := DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	return document, digest
}

func terminalEvidenceFixture(leaseSHA string) cutoverv1.TerminalEvidence {
	digest := strings.Repeat("a", 64)
	finalLSN := "0/20"
	initial := cutoverv1.TerminalInitialCopyEvidence{
		ManifestSHA256: leaseSHA, SnapshotName: "snapshot-main", SnapshotLSN: "0/10",
		RelationCount: 1, PartitionCount: 1, RowCount: 1, RootSHA256: digest,
	}
	checkpointCount, journalCount, appliedCount := 1, int64(1), int64(1)
	return cutoverv1.TerminalEvidence{
		ContractVersion: "cutover-terminal-evidence.v1",
		OperationID:     attestationOperationID, RunID: attestationRunID,
		LeaseManifestSHA256: leaseSHA, PlanSHA256: digest,
		TransformManifestSHA256: digest, TerminalState: "completed", FinalLSN: &finalLSN,
		InitialCopy: &initial, CheckpointCount: &checkpointCount,
		Checkpoints: []cutoverv1.TerminalCheckpointEvidence{{
			RelationOrdinal: 1, CheckpointKey: "main", AppliedLSN: "0/20",
			AppliedRows: 1, BatchCount: 1, BatchRootSHA256: digest,
		}},
		JournalTransactionCount: &journalCount, AppliedChangeCount: &appliedCount,
		JournalCoverageRootSHA256: &digest,
	}
}
