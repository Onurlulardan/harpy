package systemtransfer

import (
	"crypto/sha256"
	"testing"
	"time"
)

func TestExportArtifactReconciliationBindsExactEffectEvidence(t *testing.T) {
	now := time.Date(2026, time.July, 18, 9, 0, 0, 123000000, time.UTC)
	attempt := fixtureExportAttempt(t, now)
	attempt.AttemptID = fixtureAttemptID
	attempt.LeaseOwnerID = fixtureLeaseID
	attempt.AttemptGeneration = 1
	effect := ExportArtifactEffectStart{Attempt: attempt,
		EffectID: attempt.AttemptID, EffectReceiptID: fixtureExportReceiptID,
		ObjectKey:     "iam/exports/artifact.fcart",
		RequestSHA256: sha256.Sum256([]byte("provider-request")), StartedAt: now,
		Draft: ExportArtifactDraftEvidence{KeyID: "iam-export-key-1",
			HeaderSHA256: [32]byte{1}, CiphertextSHA256: [32]byte{2},
			CiphertextBytes: 128, WrappedDEK: make([]byte, 32),
			WrapNonce: [12]byte{1}}}
	artifact := ExportArtifactMetadata{ObjectKey: effect.ObjectKey,
		ObjectVersion: "version-1", KeyID: "iam-export-key-1",
		ProviderRequestID: "provider-request-1", EffectID: effect.EffectID,
		EffectReceiptID: effect.EffectReceiptID, HeaderSHA256: [32]byte{1},
		CiphertextSHA256: [32]byte{2}, RequestSHA256: effect.RequestSHA256,
		ResultSHA256:    sha256.Sum256([]byte("provider-result")),
		CiphertextBytes: 128, WrappedDEK: make([]byte, 32), WrapNonce: [12]byte{1},
		AttemptGeneration: 1, SealedAt: now}
	result := ExportArtifactReconciliation{State: ExportArtifactSealed,
		Artifact: artifact, ProviderRequestID: "reconcile-provider-1",
		EvidenceSHA256: sha256.Sum256([]byte("reconciliation")), CheckedAt: now}
	if effect.Validate() != nil || result.Validate(effect) != nil {
		t.Fatal("exact reconciliation evidence was rejected")
	}
	result.Artifact.RequestSHA256 = sha256.Sum256([]byte("different-request"))
	if result.Validate(effect) == nil {
		t.Fatal("reconciliation accepted a different provider request")
	}
	result = ExportArtifactReconciliation{State: ExportArtifactMissing,
		ProviderRequestID: "reconcile-provider-2",
		EvidenceSHA256:    sha256.Sum256([]byte("missing")), CheckedAt: now}
	if result.Validate(effect) != nil {
		t.Fatal("verified missing result was rejected")
	}
	result.Artifact.ProviderRequestID = "unexpected-provider-evidence"
	if result.Validate(effect) == nil {
		t.Fatal("missing result accepted uploaded artifact evidence")
	}
}
