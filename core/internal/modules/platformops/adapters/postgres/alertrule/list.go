package alertrule

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (operations *AlertRuleOperations) ListAlertRules(
	ctx context.Context, environmentID string,
) ([]platformopsapi.AlertRule, error) {
	return operations.backend.AlertRuleListing(ctx, environmentID)
}
