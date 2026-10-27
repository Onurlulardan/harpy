package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func CompleteTenantCheckpoint(
	ctx context.Context,
	transaction *sql.Tx,
	identity TenantCheckpointIdentity,
	state string,
	errorCode any,
) error {
	arguments := []any{
		identity.TenantID, identity.Owner, identity.Version, identity.Attempt,
		identity.Origin, identity.SelectedPath, identity.SelectedHash, identity.RunID,
		state, errorCode,
	}
	if _, err := transaction.ExecContext(
		ctx, insertTenantCheckpointAttemptSQL, arguments...,
	); err != nil {
		return fmt.Errorf("insert tenant checkpoint attempt: %w", err)
	}
	result, err := transaction.ExecContext(ctx, completeTenantCheckpointSQL, arguments...)
	if err != nil {
		return fmt.Errorf("complete tenant checkpoint: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count completed tenant checkpoint: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("complete tenant checkpoint affected %d rows", affected)
	}
	return nil
}
