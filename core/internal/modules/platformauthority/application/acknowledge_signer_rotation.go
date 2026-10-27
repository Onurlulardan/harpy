package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

func (service *Coordinator) AcknowledgeSignerRotation(
	ctx context.Context,
	operationID string,
) (api.AuthorityReceipt, error) {
	if strings.TrimSpace(operationID) == "" {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	operation, err := service.store.LoadOperation(ctx, operationID)
	if err != nil || operation.Purpose != "proposal_signer_rotation" ||
		operation.Proposal.Value.SignerRotation == nil {
		return api.AuthorityReceipt{}, mapStoreError(errOrConflict(err))
	}
	receipt, err := service.completeOperation(ctx, operation)
	if err != nil {
		return api.AuthorityReceipt{}, err
	}
	completed, loadErr := service.store.LoadOperation(ctx, operationID)
	if loadErr != nil {
		return api.AuthorityReceipt{}, mapStoreError(loadErr)
	}
	err = service.applyCompletedRotation(ctx, completed)
	if err != nil {
		return api.AuthorityReceipt{}, mapStoreError(err)
	}
	return receipt, nil
}
