package devjournal

import (
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func (registry *Registry) createAck(
	state journalState,
	proposal artifact.SignedProposal,
	now time.Time,
) (artifact.VerifiedAck, error) {
	if !validRegistryInstant(now) {
		return artifact.VerifiedAck{}, fmt.Errorf("development authority clock is invalid")
	}
	key := headMapKey(
		proposal.Value.RegistryBindingID, proposal.Value.ScopeKind, proposal.Value.ScopeID,
	)
	current, exists := state.heads[key]
	if (proposal.Value.Generation == 1) != !exists {
		return artifact.VerifiedAck{}, fmt.Errorf("development authority generation conflicts")
	}
	if exists && (proposal.Value.Generation != current.Generation+1 ||
		proposal.Value.PreviousHeadSHA256 == nil ||
		*proposal.Value.PreviousHeadSHA256 != current.AckSHA256) {
		return artifact.VerifiedAck{}, fmt.Errorf("development authority previous head conflicts")
	}
	ack := artifact.Ack{
		Schema: artifact.AckSchema, RegistryKind: artifact.RegistryDevelopment,
		RegistryBindingID: proposal.Value.RegistryBindingID,
		ScopeKind:         proposal.Value.ScopeKind, ScopeID: proposal.Value.ScopeID,
		RegistrySequence: state.sequence + 1, OperationID: proposal.Value.OperationID,
		ProposalSHA256: proposal.SHA256, Generation: proposal.Value.Generation,
		PreviousHeadSHA256: proposal.Value.PreviousHeadSHA256,
		StateSHA256:        proposal.Value.StateSHA256, IssuedAt: now.Format(time.RFC3339),
		RegistryKID: registry.registryKID,
	}
	return artifact.SignAck(ack, proposal, registry.privateKey)
}
