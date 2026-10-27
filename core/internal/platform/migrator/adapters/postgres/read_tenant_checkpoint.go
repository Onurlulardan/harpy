package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func ReadTenantCheckpoint(
	ctx context.Context,
	database tenantRowQuerier,
	tenantID, owner string,
	version int64,
) (TenantCheckpoint, bool, error) {
	checkpoint, err := scanTenantCheckpoint(database.QueryRowContext(
		ctx, readTenantCheckpointSQL, tenantID, owner, version,
	), tenantID, owner, version)
	if errors.Is(err, sql.ErrNoRows) {
		return TenantCheckpoint{}, false, nil
	}
	if err != nil {
		return TenantCheckpoint{}, false, fmt.Errorf("read tenant checkpoint: %w", err)
	}
	return checkpoint, true, nil
}

func scanTenantCheckpoint(
	row *sql.Row,
	tenantID, owner string,
	version int64,
) (TenantCheckpoint, error) {
	checkpoint := TenantCheckpoint{TenantCheckpointIdentity: TenantCheckpointIdentity{
		TenantID: tenantID, Owner: owner, Version: version,
	}}
	err := row.Scan(
		&checkpoint.Origin, &checkpoint.SelectedPath, &checkpoint.SelectedHash,
		&checkpoint.RunID, &checkpoint.State, &checkpoint.Attempt,
		&checkpoint.StartedAt, &checkpoint.FinishedAt, &checkpoint.ErrorCode,
		&checkpoint.UpdatedAt,
	)
	return checkpoint, err
}
