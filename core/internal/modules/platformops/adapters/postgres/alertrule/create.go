package alertrule

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *AlertRuleOperations) CreateAlertRule(
	ctx context.Context, input platformopsapi.AlertRuleInput,
) (platformopsapi.AlertRule, error) {
	return operations.backend.AlertRuleCreation(ctx, input)
}
