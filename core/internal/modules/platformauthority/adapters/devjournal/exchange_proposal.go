package devjournal

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func (registry *Registry) ExchangeProposal(
	ctx context.Context,
	input artifact.SignedProposal,
) ([]byte, error) {
	var result []byte
	err := registry.withLock(ctx, func() error {
		state, err := registry.loadState(ctx)
		if err != nil {
			return err
		}
		parsed, err := artifact.ParseSignedProposal(input.Bytes)
		if err != nil {
			return fmt.Errorf("development authority proposal is invalid")
		}
		if existing, ok := state.operations[parsed.Value.OperationID]; ok {
			if existing.ack.Value.ProposalSHA256 != parsed.SHA256 {
				return fmt.Errorf("development authority operation replay conflicts")
			}
			result = append([]byte(nil), existing.ack.Bytes...)
			return nil
		}
		binding, exists := state.bindings[parsed.Value.RegistryBindingID]
		if !exists || binding.proof.Value.EnvironmentID != parsed.Value.EnvironmentID ||
			binding.signerKID != parsed.Value.SignerKID {
			return fmt.Errorf("development authority proposal binding conflicts")
		}
		key, err := registry.resolver.ResolveProposalKey(ctx, parsed.Value.SignerKID)
		if err != nil {
			return fmt.Errorf("development authority proposal key is unavailable")
		}
		now := registry.clock.Now()
		proposal, err := artifact.VerifyProposalSignature(
			input.Bytes, artifact.ExpiryOnline, now, key,
		)
		if err != nil {
			return fmt.Errorf("development authority proposal signature is invalid")
		}
		ack, err := registry.createAck(state, proposal, now)
		if err != nil {
			return err
		}
		scopeKind := string(ack.Value.ScopeKind)
		scopeID := ack.Value.ScopeID
		operationID := ack.Value.OperationID
		proposalSHA := ack.Value.ProposalSHA256
		generation := ack.Value.Generation
		stateSHA := ack.Value.StateSHA256
		record := journalRecord{
			Kind: "ack", RegistrySequence: ack.Value.RegistrySequence,
			RegistryBindingID: ack.Value.RegistryBindingID, ScopeKind: &scopeKind,
			ScopeID: &scopeID, OperationID: &operationID, ProposalSHA256: &proposalSHA,
			Generation: &generation, PreviousHeadSHA256: ack.Value.PreviousHeadSHA256,
			StateSHA256: &stateSHA,
			Artifact:    base64.RawURLEncoding.EncodeToString(ack.Bytes),
		}
		encoded := base64.RawURLEncoding.EncodeToString(proposal.Bytes)
		record.ProposalArtifact = &encoded
		if err := appendRecord(registry.path, record); err != nil {
			return err
		}
		result = append([]byte(nil), ack.Bytes...)
		return nil
	})
	return result, err
}

func validRegistryInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}
