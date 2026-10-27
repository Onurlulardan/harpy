package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

func (service *AirGapCoordinator) PrepareSignerRotationExport(
	ctx context.Context,
	request api.SignerRotationRequest,
) (api.AirGapRequestArtifact, error) {
	if request.Policy != api.ExpiryAirGap {
		return api.AirGapRequestArtifact{}, api.ErrConflict
	}
	prepared, err := service.base.PrepareSignerRotation(ctx, request)
	if err != nil {
		return api.AirGapRequestArtifact{}, err
	}
	operation, err := service.base.store.LoadOperation(ctx, prepared.OperationID)
	if err != nil {
		return api.AirGapRequestArtifact{}, mapStoreError(err)
	}
	return airGapArtifact(operation), nil
}
