package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func StartTenantCheckpoint(
	ctx context.Context,
	transaction *sql.Tx,
	identity TenantCheckpointIdentity,
	previous TenantCheckpoint,
) error {
	result, err := transaction.ExecContext(
		ctx, startTenantCheckpointSQL,
		identity.TenantID, identity.Owner, identity.Version,
		identity.RunID, identity.Attempt,
		identity.Origin, identity.SelectedPath, identity.SelectedHash,
		previous.Origin, previous.SelectedPath, previous.SelectedHash,
		previous.State, previous.Attempt, previous.RunID,
	)
	if err != nil {
		return fmt.Errorf("start tenant checkpoint: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count started tenant checkpoint: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("start tenant checkpoint affected %d rows", affected)
	}
	return nil
}
