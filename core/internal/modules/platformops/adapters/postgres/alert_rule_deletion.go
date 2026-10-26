package postgres

import (
	"context"
	"fmt"
)

func (store *Store) AlertRuleDeletion(
	ctx context.Context,
	environmentID, id string,
) error {
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		_, err := database.Execute(ctx, deleteAlertRuleSQL, id, environmentID)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("delete alert rule: %w", err)
	}
	return nil
}
