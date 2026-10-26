package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

func (service *AirGapCoordinator) PrepareChangeExport(
	ctx context.Context,
	request api.ProtectedChangeRequest,
) (api.AirGapRequestArtifact, error) {
	if request.Policy != api.ExpiryAirGap {
		return api.AirGapRequestArtifact{}, api.ErrConflict
	}
	prepared, err := service.base.PrepareChange(ctx, request)
	if err != nil {
		return api.AirGapRequestArtifact{}, err
	}
	operation, err := service.base.store.LoadOperation(ctx, prepared.OperationID)
	if err != nil {
		return api.AirGapRequestArtifact{}, mapStoreError(err)
	}
	return airGapArtifact(operation), nil
}
