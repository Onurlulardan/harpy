package attestation

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func VerifyCutoverTerminalEvidence(
	document []byte,
	adoption VerifiedCutoverCredentialLease,
	finalizer VerifiedCutoverFinalizerLease,
) (VerifiedCutoverTerminalEvidence, error) {
	if !adoption.verified || !finalizer.verified {
		return VerifiedCutoverTerminalEvidence{}, fmt.Errorf("cutover terminal prerequisites were not verified")
	}
	var evidence cutoverv1.TerminalEvidence
	if err := decodeStrictBounded(document, &evidence,
		cutoverv1.MaximumTerminalEvidenceBytes); err != nil {
		return VerifiedCutoverTerminalEvidence{}, fmt.Errorf("decode cutover terminal evidence: %w", err)
	}
	if err := cutoverv1.ValidateTerminalEvidence(evidence); err != nil {
		return VerifiedCutoverTerminalEvidence{}, err
	}
	canonical, err := canonicalizeBounded(document, cutoverv1.MaximumTerminalEvidenceBytes)
	if err != nil || !bytes.Equal(document, canonical) {
		return VerifiedCutoverTerminalEvidence{}, fmt.Errorf("cutover terminal evidence is not canonical")
	}
	digestBytes := sha256.Sum256(canonical)
	digest := hex.EncodeToString(digestBytes[:])
	if subtle.ConstantTimeCompare([]byte(digest),
		[]byte(finalizer.TerminalEvidenceSHA256())) != 1 {
		return VerifiedCutoverTerminalEvidence{}, fmt.Errorf("cutover terminal evidence digest differs")
	}
	if err := bindTerminalEvidence(evidence, adoption, finalizer.TerminalState()); err != nil {
		return VerifiedCutoverTerminalEvidence{}, err
	}
	return VerifiedCutoverTerminalEvidence{
		evidence: evidence, digest: digest, canonical: canonical, verified: true,
	}, nil
}

func bindTerminalEvidence(
	evidence cutoverv1.TerminalEvidence,
	adoption VerifiedCutoverCredentialLease,
	expectedState string,
) error {
	if evidence.OperationID != adoption.OperationID() || evidence.RunID != adoption.RunID() ||
		evidence.LeaseManifestSHA256 != adoption.ManifestSHA256() ||
		evidence.PlanSHA256 != adoption.PlanSHA256() ||
		evidence.TransformManifestSHA256 != adoption.TransformManifestSHA256() ||
		evidence.TerminalState != expectedState {
		return fmt.Errorf("cutover terminal evidence authority differs")
	}
	if evidence.TerminalState != "completed" {
		return nil
	}
	if evidence.InitialCopy == nil ||
		evidence.InitialCopy.RelationCount != adoption.RelationCount() {
		return fmt.Errorf("cutover terminal evidence authority differs")
	}
	checkpointIndex := 0
	for ordinal := 1; ordinal <= adoption.RelationCount(); ordinal++ {
		relation, found := adoption.Relation(ordinal)
		if !found {
			return fmt.Errorf("cutover terminal relation plan differs")
		}
		for _, expectedKey := range relation.ExpectedCheckpointKeys {
			if checkpointIndex >= len(evidence.Checkpoints) {
				return fmt.Errorf("cutover terminal checkpoint set differs")
			}
			checkpoint := evidence.Checkpoints[checkpointIndex]
			if checkpoint.RelationOrdinal != ordinal || checkpoint.CheckpointKey != expectedKey {
				return fmt.Errorf("cutover terminal checkpoint set differs")
			}
			checkpointIndex++
		}
	}
	if checkpointIndex != len(evidence.Checkpoints) {
		return fmt.Errorf("cutover terminal checkpoint set differs")
	}
	return nil
}
