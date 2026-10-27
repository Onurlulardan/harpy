package alertrule

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *AlertRuleOperations) UpdateAlertRule(
	ctx context.Context, id string, input platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	return operations.backend.AlertRuleUpdate(ctx, id, input)
}
