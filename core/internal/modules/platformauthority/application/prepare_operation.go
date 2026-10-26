package application

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) prepareOperation(
	ctx context.Context,
	command operationCommand,
) (ports.OperationRecord, error) {
	existing, err := service.store.LoadOperation(ctx, command.operationID)
	if err == nil {
		if !operationMatchesCommand(existing, command) {
			return ports.OperationRecord{}, api.ErrConflict
		}
		return existing, nil
	}
	rotation := command.purpose == "proposal_signer_rotation" && command.rotation != nil
	if !errors.Is(err, ports.ErrNotFound) || !validIssuedAt(command.issuedAt) ||
		(!rotation && command.state.Value.ProposalSignerKID != service.signer.KID()) ||
		(rotation && command.state.Value.ProposalSignerKID == service.signer.KID()) {
		return ports.OperationRecord{}, api.ErrConflict
	}
	publicKey, err := service.activeSignerPublicKey(ctx, command)
	if err != nil {
		return ports.OperationRecord{}, err
	}
	value := artifact.Proposal{
		Schema: artifact.ProposalSchema, Purpose: command.purpose,
		OperationID: command.operationID, EnvironmentID: command.state.Value.EnvironmentID,
		DeploymentMode: command.state.Value.DeploymentMode,
		ScopeKind:      command.state.Value.ScopeKind, ScopeID: command.state.Value.ScopeID,
		RegistryBindingID: command.state.Value.RegistryBindingID,
		Generation:        command.generation, PreviousHeadSHA256: command.previousHead,
		StateSHA256:               command.state.SHA256,
		RecoveryTrustBundleSHA256: command.state.Value.RecoveryTrustBundleSHA256,
		CompensatesAckSHA256:      command.compensatesAck,
		SignerRotation:            command.rotation,
		IssuedAt:                  command.issuedAt.Format(time.RFC3339),
		ExpiresAt:                 command.issuedAt.Add(expiryFor(command.policy)).Format(time.RFC3339),
		SignerKID:                 service.signer.KID(),
	}
	prepared, err := artifact.PrepareProposal(value, command.state, command.policy)
	if err != nil {
		return ports.OperationRecord{}, api.ErrConflict
	}
	signature, err := service.signer.Sign(ctx, prepared.SignatureInput)
	if err != nil {
		return ports.OperationRecord{}, api.ErrDependency
	}
	proposal, err := artifact.AttachProposalSignature(prepared, signature)
	if err != nil {
		return ports.OperationRecord{}, api.ErrDependency
	}
	if _, err := artifact.VerifyProposal(
		proposal.Bytes, command.state, command.policy, command.issuedAt, publicKey,
	); err != nil {
		return ports.OperationRecord{}, api.ErrDependency
	}
	operation, err := service.store.PrepareOperation(ctx, ports.OperationRecord{
		OperationID: command.operationID, Purpose: command.purpose,
		State: command.state, Proposal: proposal, Status: ports.StatusProposalSigned,
	})
	if err != nil {
		return ports.OperationRecord{}, mapStoreError(err)
	}
	return operation, nil
}

func validIssuedAt(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}

func operationMatchesCommand(existing ports.OperationRecord, command operationCommand) bool {
	value := existing.Proposal.Value
	return existing.OperationID == command.operationID && existing.Purpose == command.purpose &&
		existing.State.SHA256 == command.state.SHA256 && value.Generation == command.generation &&
		optionalStringEqual(value.PreviousHeadSHA256, command.previousHead) &&
		optionalStringEqual(value.CompensatesAckSHA256, command.compensatesAck) &&
		rotationEvidenceEqual(value.SignerRotation, command.rotation) &&
		value.IssuedAt == command.issuedAt.Format(time.RFC3339) &&
		value.ExpiresAt == command.issuedAt.Add(expiryFor(command.policy)).Format(time.RFC3339)
}

func rotationEvidenceEqual(left, right *artifact.SignerRotationEvidence) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.PreviousSignerGeneration == right.PreviousSignerGeneration &&
		left.NewSignerKID == right.NewSignerKID &&
		left.NewSignerPublicKeySHA256 == right.NewSignerPublicKeySHA256 &&
		optionalStringEqual(left.RecoveryTrustBundleBytes, right.RecoveryTrustBundleBytes)
}

func optionalStringEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
