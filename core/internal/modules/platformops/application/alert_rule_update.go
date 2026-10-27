package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) AlertRuleUpdate(
	ctx context.Context,
	id string,
	input platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	if service == nil || service.alerts == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(id) || !validAlertInput(input) {
		return platformopsapi.AlertRule{}, platformopsapi.ErrInvalid
	}
	result, err := service.alerts.UpdateAlertRule(ctx, id, input)
	if err != nil {
		return platformopsapi.AlertRule{}, mapRecoveryStoreError(err)
	}
	return result, nil
}
