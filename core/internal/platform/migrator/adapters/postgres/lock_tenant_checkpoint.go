package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func LockTenantCheckpoint(
	ctx context.Context,
	transaction *sql.Tx,
	tenantID, owner string,
	version int64,
) (TenantCheckpoint, bool, error) {
	checkpoint, err := scanTenantCheckpoint(transaction.QueryRowContext(
		ctx, lockTenantCheckpointSQL, tenantID, owner, version,
	), tenantID, owner, version)
	if errors.Is(err, sql.ErrNoRows) {
		return TenantCheckpoint{}, false, nil
	}
	if err != nil {
		return TenantCheckpoint{}, false, fmt.Errorf("lock tenant checkpoint: %w", err)
	}
	return checkpoint, true, nil
}
