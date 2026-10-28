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
