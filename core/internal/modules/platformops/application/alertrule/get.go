package alertrule

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *AlertRuleOperations) GetAlertRule(
	ctx context.Context, environmentID, ruleID string,
) (platformopsapi.AlertRule, error) {
	return operations.backend.AlertRuleLookup(ctx, environmentID, ruleID)
}
