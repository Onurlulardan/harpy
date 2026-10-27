package restoreexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) change(
	ctx context.Context,
	statement string,
	identity ports.RestoreExecutionIdentity,
	arguments ...any,
) (bool, error) {
	var changed bool
	values := []any{identity.OperationID, identity.AttemptID, identity.LeaseToken}
	values = append(values, arguments...)
	err := store.withBound(ctx, func(database boundDatabase) error {
		return database.QueryRow(ctx, statement, values...).Scan(&changed)
	})
	if err != nil {
		return false, fmt.Errorf("change restore execution: %w", err)
	}
	return changed, nil
}
