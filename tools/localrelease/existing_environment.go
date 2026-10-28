package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func readExistingEnvironmentID(
	ctx context.Context,
	target *pgx.Conn,
	live liveDatabaseIdentity,
) (string, error) {
	var exists bool
	if err := target.QueryRow(ctx,
		"SELECT pg_catalog.to_regclass('platform.environment_identity') IS NOT NULL",
	).Scan(&exists); err != nil {
		return "", fmt.Errorf("inspect existing environment identity: %w", err)
	}
	if !exists {
		return "", nil
	}
	var environmentID, origin, systemID, releaseID, authorityDigest string
	var databaseOID uint32
	err := target.QueryRow(ctx, `SELECT environment_id::text, database_origin,
database_system_identifier, database_oid, release_id::text,
bootstrap_authority_sha256 FROM platform.environment_identity WHERE singleton=TRUE`).Scan(
		&environmentID, &origin, &systemID, &databaseOID, &releaseID, &authorityDigest,
	)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read existing environment identity: %w", err)
	}
	if _, err := uuid.Parse(environmentID); err != nil {
		return "", fmt.Errorf("existing environment ID is invalid")
	}
	if _, err := uuid.Parse(releaseID); err != nil {
		return "", fmt.Errorf("existing bootstrap release ID is invalid")
	}
	digest, err := hex.DecodeString(authorityDigest)
	if err != nil || len(digest) != 32 || origin != "fresh" ||
		systemID != live.systemID || databaseOID != live.databaseOID {
		return "", fmt.Errorf("existing environment identity differs from the live database")
	}
	return environmentID, nil
}

func readAppliedReleaseID(
	ctx context.Context, target *pgx.Conn, environmentID, digest string,
) (string, error) {
	if environmentID == "" || digest == "" {
		return "", nil
	}
	var exists bool
	if err := target.QueryRow(ctx,
		"SELECT pg_catalog.to_regclass('platform.release_generations') IS NOT NULL",
	).Scan(&exists); err != nil {
		return "", fmt.Errorf("inspect applied release: %w", err)
	}
	if !exists {
		return "", nil
	}
	var releaseID string
	err := target.QueryRow(ctx, `SELECT release_id::text
FROM platform.release_generations
WHERE environment_id=$1::uuid AND migration_manifest_sha256=$2
ORDER BY database_generation DESC LIMIT 1`, environmentID, digest).Scan(&releaseID)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read applied release: %w", err)
	}
	if _, err := uuid.Parse(releaseID); err != nil {
		return "", fmt.Errorf("applied release id is invalid")
	}
	return releaseID, nil
}
