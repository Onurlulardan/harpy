package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) AlertRuleListing(
	ctx context.Context,
	environmentID string,
) ([]platformopsapi.AlertRule, error) {
	if service == nil || service.alerts == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) {
		return nil, platformopsapi.ErrInvalid
	}
	result, err := service.alerts.ListAlertRules(ctx, environmentID)
	if err != nil {
		return nil, mapRecoveryStoreError(err)
	}
	return result, nil
}
