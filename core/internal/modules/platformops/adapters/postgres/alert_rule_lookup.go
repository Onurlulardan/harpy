package postgres

import (
	"context"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (store *Store) AlertRuleLookup(
	ctx context.Context, environmentID, ruleID string,
) (platformopsapi.AlertRule, error) {
	var result platformopsapi.AlertRule
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		value, err := scanAlertRule(database.QueryRow(
			ctx, getAlertRuleSQL, environmentID, ruleID))
		if err != nil {
			return store.mapDatabaseError(err)
		}
		result = value
		return nil
	})
	if err != nil {
		return platformopsapi.AlertRule{}, fmt.Errorf("get alert rule: %w", err)
	}
	return result, nil
}
