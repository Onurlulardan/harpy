package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	baselineSeedEnvironmentID = "018f0000-0000-7000-8000-000000000102"
	baselineSeedReleaseID     = "018f0000-0000-7000-8000-000000000101"
	baselineSeedAuthority     = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	baselineSeedSystemID      = "7703580167595228103"
	baselineSeedDatabaseOID   = int64(51019)
)

func bindBaselineSeedIdentity(
	ctx context.Context, target *pgx.Conn, live liveDatabaseIdentity,
) error {
	var exists bool
	if err := target.QueryRow(ctx,
		"SELECT pg_catalog.to_regclass('platform.environment_identity') IS NOT NULL",
	).Scan(&exists); err != nil {
		return fmt.Errorf("inspect baseline seed identity: %w", err)
	}
	if !exists {
		return nil
	}
	_, err := target.Exec(ctx, `UPDATE platform.environment_identity
SET database_system_identifier=$1, database_oid=$2
WHERE singleton=TRUE AND database_origin='fresh'
  AND environment_id=$3::uuid AND release_id=$4::uuid
  AND bootstrap_authority_sha256=$5
  AND database_system_identifier=$6 AND database_oid=$7`,
		live.systemID, int64(live.databaseOID),
		baselineSeedEnvironmentID, baselineSeedReleaseID, baselineSeedAuthority,
		baselineSeedSystemID, baselineSeedDatabaseOID,
	)
	if err != nil {
		return fmt.Errorf("bind baseline seed identity to the live database: %w", err)
	}
	return nil
}

func baselineSeedAwaitingRelease(
	ctx context.Context, target *pgx.Conn, environmentID string,
) (bool, error) {
	var origin, releaseID, authority string
	err := target.QueryRow(ctx, `SELECT database_origin, release_id::text,
bootstrap_authority_sha256 FROM platform.environment_identity
WHERE singleton=TRUE AND environment_id=$1::uuid`, environmentID).Scan(
		&origin, &releaseID, &authority,
	)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect baseline seed release: %w", err)
	}
	return unboundBaselineSeed(environmentID, releaseID, origin, authority), nil
}

func unboundBaselineSeed(environmentID, releaseID, origin, authority string) bool {
	return environmentID == baselineSeedEnvironmentID &&
		releaseID == baselineSeedReleaseID &&
		origin == "fresh" &&
		authority == baselineSeedAuthority
}
