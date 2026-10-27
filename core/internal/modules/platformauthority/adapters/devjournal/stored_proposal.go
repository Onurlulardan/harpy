package devjournal

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func (registry *Registry) verifyStoredProposal(
	ctx context.Context,
	state *journalState,
	record journalRecord,
	ack artifact.VerifiedAck,
) (artifact.SignedProposal, error) {
	document, err := base64.RawURLEncoding.DecodeString(*record.ProposalArtifact)
	if err != nil || base64.RawURLEncoding.EncodeToString(document) != *record.ProposalArtifact {
		return artifact.SignedProposal{}, fmt.Errorf("development authority proposal record is invalid")
	}
	binding := state.bindings[record.RegistryBindingID]
	key, err := registry.resolver.ResolveProposalKey(ctx, binding.signerKID)
	if err != nil {
		return artifact.SignedProposal{}, fmt.Errorf("development authority proposal key is unavailable")
	}
	proposal, err := artifact.VerifyStoredProposalSignature(document, key)
	if err != nil || proposal.SHA256 != *record.ProposalSHA256 ||
		proposal.Value.SignerKID != binding.signerKID ||
		proposal.Value.RegistryBindingID != record.RegistryBindingID ||
		proposal.Value.EnvironmentID != binding.proof.Value.EnvironmentID ||
		proposal.Value.OperationID != *record.OperationID ||
		proposal.Value.ScopeKind != ack.Value.ScopeKind ||
		proposal.Value.ScopeID != ack.Value.ScopeID ||
		proposal.Value.Generation != ack.Value.Generation ||
		proposal.Value.StateSHA256 != ack.Value.StateSHA256 {
		return artifact.SignedProposal{}, fmt.Errorf("development authority proposal record conflicts")
	}
	return proposal, nil
}
