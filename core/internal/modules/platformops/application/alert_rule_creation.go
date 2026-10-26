package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) AlertRuleCreation(
	ctx context.Context,
	input platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	if service == nil || service.alerts == nil || ctx == nil || ctx.Err() != nil ||
		!validAlertInput(input) {
		return platformopsapi.AlertRule{}, platformopsapi.ErrInvalid
	}
	result, err := service.alerts.CreateAlertRule(ctx, input)
	if err != nil {
		return platformopsapi.AlertRule{}, mapRecoveryStoreError(err)
	}
	return result, nil
}
