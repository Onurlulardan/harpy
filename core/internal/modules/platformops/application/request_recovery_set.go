package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *RecoveryService) RequestRecoverySet(
	ctx context.Context,
	request api.RecoverySetRequest,
) (api.RecoverySet, error) {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(request.RecoverySetID) ||
		!validRecoveryUUID(request.OperationID) ||
		!validRecoveryUUID(request.EnvironmentID) {
		return api.RecoverySet{}, api.ErrInvalid
	}
	current, err := service.identities.CurrentIdentity(ctx, request.EnvironmentID)
	if err != nil {
		return api.RecoverySet{}, mapRecoveryStoreError(err)
	}
	identity := current.Identity
	if recoveryset.ValidateIdentity(identity) != nil ||
		identity.EnvironmentID != request.EnvironmentID ||
		(recoveryset.ModeSelection{Mode: identity.DeploymentMode,
			OnPremiseStorage: true}).Validate() != nil {
		return api.RecoverySet{}, api.ErrConflict
	}
	record := ports.RecoverySetRequestRecord{RecoverySetID: request.RecoverySetID,
		OperationID: request.OperationID, EnvironmentID: identity.EnvironmentID,
		CreatedAt: service.clock.Now()}
	sequence, err := service.sets.RequestRecoverySet(ctx, record)
	if err != nil {
		return api.RecoverySet{}, mapRecoveryStoreError(err)
	}
	return api.RecoverySet{RecoverySetID: request.RecoverySetID,
		OperationID: request.OperationID, EnvironmentID: identity.EnvironmentID,
		Sequence: sequence, State: "requested"}, nil
}
