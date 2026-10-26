package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func (service *AirGapCoordinator) ImportAcknowledgement(
	ctx context.Context,
	request api.AirGapAckImport,
) (api.AuthorityReceipt, error) {
	operation, err := service.base.store.LoadOperation(ctx, request.OperationID)
	if err != nil {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	if operation.Status == "completed" {
		if string(operation.Ack.Bytes) != string(request.AckBytes) {
			return api.AuthorityReceipt{}, api.ErrConflict
		}
		return service.airGapResult(ctx, operation)
	}
	if operation.Status == "ack_verified" {
		committed, commitErr := service.base.store.CommitAcknowledged(ctx, operation.OperationID)
		if commitErr != nil {
			return api.AuthorityReceipt{}, mapStoreError(commitErr)
		}
		return service.airGapResult(ctx, committed.Operation)
	}
	if operation.Status != "proposal_signed" {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	ack, err := artifact.VerifyAck(
		request.AckBytes, operation.Proposal, service.base.trust, service.base.clock.Now(),
	)
	if err != nil {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	if err := service.authorizer.AuthorizeAckImport(
		ctx, operation.OperationID, ack.SHA256,
	); err != nil {
		return api.AuthorityReceipt{}, api.ErrConflict
	}
	if _, err := service.base.store.PersistAcknowledgement(
		ctx, operation.OperationID, ack,
	); err != nil {
		return api.AuthorityReceipt{}, mapStoreError(err)
	}
	committed, err := service.base.store.CommitAcknowledged(ctx, operation.OperationID)
	if err != nil {
		return api.AuthorityReceipt{}, mapStoreError(err)
	}
	return service.airGapResult(ctx, committed.Operation)
}
