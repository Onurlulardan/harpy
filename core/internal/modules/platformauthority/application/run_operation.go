package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) runOperation(
	ctx context.Context,
	command operationCommand,
) (api.AuthorityReceipt, error) {
	operation, err := service.prepareOperation(ctx, command)
	if err != nil {
		return api.AuthorityReceipt{}, err
	}
	return service.completeOperation(ctx, operation)
}

func (service *Coordinator) completeOperation(
	ctx context.Context,
	operation ports.OperationRecord,
) (api.AuthorityReceipt, error) {
	if operation.Status == ports.StatusCompleted {
		return receiptFrom(operation), nil
	}
	if operation.Status == ports.StatusProposalSigned {
		ackBytes, err := service.registry.ExchangeProposal(ctx, operation.Proposal)
		if err != nil {
			return api.AuthorityReceipt{}, api.ErrDependency
		}
		ack, err := artifact.VerifyAck(
			ackBytes, operation.Proposal, service.trust, service.clock.Now(),
		)
		if err != nil {
			return api.AuthorityReceipt{}, api.ErrDependency
		}
		operation, err = service.store.PersistAcknowledgement(ctx, operation.OperationID, ack)
		if err != nil {
			return api.AuthorityReceipt{}, mapStoreError(err)
		}
	}
	committed, err := service.store.CommitAcknowledged(ctx, operation.OperationID)
	if err != nil {
		return api.AuthorityReceipt{}, mapStoreError(err)
	}
	return receiptFrom(committed.Operation), nil
}
