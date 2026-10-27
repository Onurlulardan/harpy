package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) applyCompletedRotation(
	ctx context.Context,
	operation ports.OperationRecord,
) error {
	if operation.Purpose != "proposal_signer_rotation" {
		return nil
	}
	rotation := operation.Proposal.Value.SignerRotation
	if rotation == nil || operation.Status != ports.StatusCompleted {
		return ports.ErrConflict
	}
	return service.store.ApplySignerRotation(ctx, ports.SignerRotationRecord{
		OperationID:                  operation.OperationID,
		BindingID:                    operation.Proposal.Value.RegistryBindingID,
		ExpectedSignerGeneration:     rotation.PreviousSignerGeneration,
		ExpectedSignerKID:            operation.Proposal.Value.SignerKID,
		NewSignerGeneration:          operation.State.Value.ProposalSignerGeneration,
		NewSignerKID:                 rotation.NewSignerKID,
		NewSignerPublicKeySHA256:     rotation.NewSignerPublicKeySHA256,
		NewRecoveryTrustBundleSHA256: operation.State.Value.RecoveryTrustBundleSHA256,
	})
}
