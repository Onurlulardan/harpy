package alertrule

import "context"

func (operations *AlertRuleOperations) DeleteAlertRule(
	ctx context.Context, environmentID, id string,
) error {
	return operations.backend.AlertRuleDeletion(ctx, environmentID, id)
}
