package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func InsertTenantCheckpoint(
	ctx context.Context,
	transaction *sql.Tx,
	identity TenantCheckpointIdentity,
) error {
	_, err := transaction.ExecContext(
		ctx, insertTenantCheckpointSQL,
		identity.TenantID, identity.Owner, identity.Version, identity.Origin,
		identity.SelectedPath, identity.SelectedHash, identity.RunID,
	)
	if err != nil {
		return fmt.Errorf("insert tenant checkpoint: %w", err)
	}
	return nil
}
