package application

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (service *OperationsService) AlertRuleLookup(
	ctx context.Context, environmentID, ruleID string,
) (platformopsapi.AlertRule, error) {
	if service == nil || service.alerts == nil || ctx == nil || ctx.Err() != nil ||
		!validRecoveryUUID(environmentID) || !validRecoveryUUID(ruleID) {
		return platformopsapi.AlertRule{}, platformopsapi.ErrInvalid
	}
	result, err := service.alerts.GetAlertRule(ctx, environmentID, ruleID)
	if err != nil {
		return platformopsapi.AlertRule{}, mapRecoveryStoreError(err)
	}
	return result, nil
}
