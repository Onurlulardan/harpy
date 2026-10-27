package devjournal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func (registry *Registry) applyBindingRecord(
	ctx context.Context,
	state *journalState,
	record journalRecord,
	document []byte,
) error {
	if record.RegistrySequence != 0 || record.EnvironmentID == nil ||
		record.ScopeKind != nil || record.ScopeID != nil || record.OperationID != nil ||
		record.ProposalSHA256 != nil || record.Generation != nil ||
		record.PreviousHeadSHA256 != nil || record.StateSHA256 != nil ||
		record.ProposalArtifact != nil {
		return fmt.Errorf("development authority binding record is invalid")
	}
	// Decode once only to choose the pinned resolver key; verification remains authoritative.
	proofValue, err := decodeBindingProof(document)
	if err != nil {
		return err
	}
	key, err := registry.resolver.ResolveProposalKey(ctx, proofValue.ControlKID)
	if err != nil {
		return fmt.Errorf("development authority proposal key is unavailable")
	}
	proof, err := artifact.VerifyBindingProof(document, key)
	if err != nil || proof.Value.RegistryNonce == "" ||
		proof.Value.EnvironmentID != *record.EnvironmentID ||
		proof.Value.LicenseBindingID != record.RegistryBindingID {
		return fmt.Errorf("development authority binding proof is invalid")
	}
	if _, exists := state.bindings[record.RegistryBindingID]; exists ||
		state.environments[*record.EnvironmentID] != "" {
		return fmt.Errorf("development authority binding was rebound")
	}
	publicDigest := sha256.Sum256(key)
	state.bindings[record.RegistryBindingID] = journalBinding{proof: proof,
		signerGeneration: 1, signerKID: proof.Value.ControlKID,
		signerPublicKeySHA256: hex.EncodeToString(publicDigest[:])}
	state.environments[*record.EnvironmentID] = record.RegistryBindingID
	return nil
}
