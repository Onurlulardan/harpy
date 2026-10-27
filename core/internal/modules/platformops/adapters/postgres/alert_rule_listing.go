package postgres

import (
	"context"
	"fmt"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (store *Store) AlertRuleListing(
	ctx context.Context,
	environmentID string,
) ([]platformopsapi.AlertRule, error) {
	result := make([]platformopsapi.AlertRule, 0)
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		rows, err := database.Query(ctx, listAlertRulesSQL, environmentID)
		if err != nil {
			return store.mapDatabaseError(err)
		}
		defer rows.Close()
		for rows.Next() {
			item, scanErr := scanAlertRule(rows)
			if scanErr != nil {
				return store.mapDatabaseError(scanErr)
			}
			result = append(result, item)
		}
		return store.mapDatabaseError(rows.Err())
	})
	if err != nil {
		return nil, fmt.Errorf("list alert rules: %w", err)
	}
	return result, nil
}
