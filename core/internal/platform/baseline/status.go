package baseline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PlatformMigration struct {
	Version int64
	Path    string
	SHA256  string
}

func PlatformMigrations() []PlatformMigration {
	return []PlatformMigration{
		{Version: platformVersion, Path: "core/internal/platform/baseline/sql/" + platformFile, SHA256: digest(platformScript)},
		{Version: platformCatalogRemovalVersion, Path: "core/internal/platform/baseline/sql/" + platformCatalogRemovalFile, SHA256: digest(platformCatalogRemovalScript)},
		{Version: platformTenantExecutorBaselineVersion, Path: "core/internal/platform/baseline/sql/" + platformTenantExecutorBaselineFile, SHA256: digest(platformTenantExecutorBaselineScript)},
		{Version: platformTenantExecutorCompletionVersion, Path: "core/internal/platform/baseline/sql/" + platformTenantExecutorCompletionFile, SHA256: digest(platformTenantExecutorCompletionScript)},
		{Version: platformTenantExecutorClaimClockVersion, Path: "core/internal/platform/baseline/sql/" + platformTenantExecutorClaimClockFile, SHA256: digest(platformTenantExecutorClaimClockScript)},
		{Version: platformTenantExecutorCompletionContractVersion, Path: "core/internal/platform/baseline/sql/" + platformTenantExecutorCompletionContractFile, SHA256: digest(platformTenantExecutorCompletionContractScript)},
	}
}

func LatestPlatformVersion() int64 {
	migrations := PlatformMigrations()
	return migrations[len(migrations)-1].Version
}

func AppliedPlatformVersions(ctx context.Context, dsn string) ([]int64, error) {
	config, err := sessionConfig(dsn, "")
	if err != nil {
		return nil, err
	}
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open baseline database: %w", err)
	}
	defer connection.Close(ctx)
	var historyExists bool
	if err := connection.QueryRow(ctx, "SELECT pg_catalog.to_regclass('platform.goose_db_version') IS NOT NULL").Scan(&historyExists); err != nil {
		return nil, fmt.Errorf("inspect platform Goose history: %w", err)
	}
	applied := []int64{}
	if !historyExists {
		return applied, nil
	}
	versions := make([]int64, 0, len(PlatformMigrations()))
	for _, migration := range PlatformMigrations() {
		versions = append(versions, migration.Version)
	}
	rows, err := connection.Query(ctx,
		"SELECT version_id FROM platform.goose_db_version WHERE is_applied "+
			"AND version_id = ANY($1::bigint[]) ORDER BY version_id", versions)
	if err != nil {
		return nil, fmt.Errorf("read platform Goose history: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan platform Goose history: %w", err)
		}
		applied = append(applied, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platform Goose history: %w", err)
	}
	return applied, nil
}

func Inspect(ctx context.Context, dsn string) (bool, error) {
	applied, err := AppliedPlatformVersions(ctx, dsn)
	if err != nil {
		return false, err
	}
	return len(applied) == len(PlatformMigrations()), nil
}

func PlatformDigest() string {
	return digest(platformScript)
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
