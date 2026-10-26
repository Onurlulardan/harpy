package main

import (
	"bytes"
	"fmt"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type issuedTerminalEvidence struct {
	value    cutoverv1.TerminalEvidence
	document []byte
	digest   string
}

func loadIssuedTerminalEvidence(
	path string, adoption attestation.VerifiedCutoverCredentialLease,
) (issuedTerminalEvidence, error) {
	document, err := readArtifact(path, cutoverv1.MaximumTerminalEvidenceBytes)
	if err != nil {
		return issuedTerminalEvidence{}, err
	}
	var value cutoverv1.TerminalEvidence
	if err := attestation.DecodeStrict(document, &value); err != nil {
		return issuedTerminalEvidence{}, fmt.Errorf("terminal evidence decoding failed")
	}
	if err := cutoverv1.ValidateTerminalEvidence(value); err != nil {
		return issuedTerminalEvidence{}, fmt.Errorf("terminal evidence validation failed")
	}
	canonical, err := attestation.Canonicalize(document)
	if err != nil || !bytes.Equal(document, canonical) {
		return issuedTerminalEvidence{}, fmt.Errorf("terminal evidence is not canonical")
	}
	if err := bindIssuedTerminalEvidence(value, adoption); err != nil {
		return issuedTerminalEvidence{}, err
	}
	digest, err := attestation.DigestHex(canonical)
	if err != nil {
		return issuedTerminalEvidence{}, fmt.Errorf("terminal evidence digest failed")
	}
	return issuedTerminalEvidence{value: value, document: canonical, digest: digest}, nil
}

func bindIssuedTerminalEvidence(
	value cutoverv1.TerminalEvidence, adoption attestation.VerifiedCutoverCredentialLease,
) error {
	if value.OperationID != adoption.OperationID() || value.RunID != adoption.RunID() ||
		value.LeaseManifestSHA256 != adoption.ManifestSHA256() ||
		value.PlanSHA256 != adoption.PlanSHA256() ||
		value.TransformManifestSHA256 != adoption.TransformManifestSHA256() {
		return fmt.Errorf("terminal evidence authority differs")
	}
	if value.TerminalState != "completed" {
		return nil
	}
	if value.InitialCopy == nil ||
		value.InitialCopy.RelationCount != adoption.RelationCount() {
		return fmt.Errorf("terminal evidence relation authority differs")
	}
	checkpoint := 0
	for ordinal := 1; ordinal <= adoption.RelationCount(); ordinal++ {
		relation, found := adoption.Relation(ordinal)
		if !found {
			return fmt.Errorf("terminal evidence relation plan differs")
		}
		for _, key := range relation.ExpectedCheckpointKeys {
			if checkpoint >= len(value.Checkpoints) ||
				value.Checkpoints[checkpoint].RelationOrdinal != ordinal ||
				value.Checkpoints[checkpoint].CheckpointKey != key {
				return fmt.Errorf("terminal evidence checkpoint authority differs")
			}
			checkpoint++
		}
	}
	if checkpoint != len(value.Checkpoints) {
		return fmt.Errorf("terminal evidence checkpoint authority differs")
	}
	return nil
}
