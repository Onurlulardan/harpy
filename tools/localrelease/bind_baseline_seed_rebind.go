package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func readBaselineIdentity(
	ctx context.Context, target *pgx.Conn,
) (baselineIdentityRow, bool, error) {
	var exists bool
	if err := target.QueryRow(ctx,
		"SELECT pg_catalog.to_regclass('platform.environment_identity') IS NOT NULL",
	).Scan(&exists); err != nil {
		return baselineIdentityRow{}, false, fmt.Errorf("inspect baseline seed identity: %w", err)
	}
	if !exists {
		return baselineIdentityRow{}, false, nil
	}
	var row baselineIdentityRow
	err := target.QueryRow(ctx, `SELECT environment_id::text, release_id::text,
database_origin, bootstrap_authority_sha256, database_system_identifier, database_oid
FROM platform.environment_identity WHERE singleton=TRUE`).Scan(
		&row.environmentID, &row.releaseID, &row.origin, &row.authority,
		&row.systemID, &row.databaseOID,
	)
	if err == pgx.ErrNoRows {
		return baselineIdentityRow{}, false, nil
	}
	if err != nil {
		return baselineIdentityRow{}, false, fmt.Errorf("read baseline seed identity: %w", err)
	}
	return row, true, nil
}

func rebindPlaceholderSeed(
	ctx context.Context, target *pgx.Conn, live liveDatabaseIdentity,
) error {
	transaction, err := target.Begin(ctx)
	if err != nil {
		return fmt.Errorf("bind baseline seed identity to the live database: %w", err)
	}
	defer transaction.Rollback(ctx)
	if err := suspendEnvironmentIdentityTrigger(ctx, transaction); err != nil {
		return err
	}
	command, err := transaction.Exec(ctx, bindBaselineSeedSQL,
		live.systemID, int64(live.databaseOID),
		baselineSeedEnvironmentID, baselineSeedReleaseID, baselineSeedAuthority,
		baselineSeedSystemID, baselineSeedDatabaseOID,
	)
	if err != nil {
		return fmt.Errorf("bind baseline seed identity to the live database: %w", err)
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("bind baseline seed identity to the live database: placeholder row was not updated")
	}
	if err := resumeEnvironmentIdentityTrigger(ctx, transaction); err != nil {
		return err
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("bind baseline seed identity to the live database: %w", err)
	}
	return nil
}
