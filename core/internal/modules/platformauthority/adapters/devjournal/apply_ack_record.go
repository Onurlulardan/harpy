package devjournal

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (registry *Registry) applyAckRecord(
	ctx context.Context,
	state *journalState,
	record journalRecord,
	document []byte,
	publicKey ed25519.PublicKey,
) error {
	if record.RegistrySequence != state.sequence+1 || record.EnvironmentID != nil ||
		record.ScopeKind == nil || record.ScopeID == nil || record.OperationID == nil ||
		record.ProposalSHA256 == nil || record.Generation == nil || record.StateSHA256 == nil ||
		record.ProposalArtifact == nil ||
		state.bindings[record.RegistryBindingID].proof.Bytes == nil {
		return fmt.Errorf("development authority acknowledgement record is invalid")
	}
	ack, err := artifact.VerifyAckSignature(
		document, artifact.RegistryDevelopment, publicKey,
	)
	if err != nil || !ackRecordMatches(record, ack) {
		return fmt.Errorf("development authority acknowledgement signature is invalid")
	}
	proposal, err := registry.verifyStoredProposal(ctx, state, record, ack)
	if err != nil {
		return err
	}
	if existing, exists := state.operations[*record.OperationID]; exists {
		if existing.ack.SHA256 != ack.SHA256 {
			return fmt.Errorf("development authority operation replay conflicts")
		}
		return nil
	}
	key := headMapKey(record.RegistryBindingID, artifact.ScopeKind(*record.ScopeKind), *record.ScopeID)
	current, exists := state.heads[key]
	if (*record.Generation == 1) != !exists ||
		(exists && (record.PreviousHeadSHA256 == nil ||
			*record.PreviousHeadSHA256 != current.AckSHA256 || *record.Generation != current.Generation+1)) {
		return fmt.Errorf("development authority head chain is invalid")
	}
	if proposal.Value.Purpose == "proposal_signer_rotation" {
		if err := registry.applySignerRotation(ctx, state, proposal); err != nil {
			return err
		}
	}
	state.sequence = record.RegistrySequence
	state.operations[*record.OperationID] = journalAck{ack: ack}
	state.heads[key] = ports.ExternalHead{
		Key: ports.HeadKey{RegistryBindingID: record.RegistryBindingID,
			ScopeKind: artifact.ScopeKind(*record.ScopeKind), ScopeID: *record.ScopeID},
		Generation: *record.Generation, StateSHA256: *record.StateSHA256,
		AckSHA256: ack.SHA256, RegistrySequence: record.RegistrySequence,
	}
	return nil
}

func ackRecordMatches(record journalRecord, ack artifact.VerifiedAck) bool {
	value := ack.Value
	return value.RegistryBindingID == record.RegistryBindingID &&
		string(value.ScopeKind) == *record.ScopeKind && value.ScopeID == *record.ScopeID &&
		value.RegistrySequence == record.RegistrySequence && value.OperationID == *record.OperationID &&
		value.ProposalSHA256 == *record.ProposalSHA256 && value.Generation == *record.Generation &&
		value.StateSHA256 == *record.StateSHA256
}
