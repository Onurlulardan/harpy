package main

import (
	"context"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

func verifyNewReleaseStage(
	ctx context.Context,
	target *pgx.Conn,
	contract runtimev1.BinarySchemaContract,
	environmentID string,
) error {
	if contract.MigrationManifestSHA256 == foundationMigrationDigest {
		if environmentID != "" {
			return fmt.Errorf("bootstrap baseline already exists; its evidence cannot be replaced")
		}
		return nil
	}
	if environmentID == "" {
		agentDebugLog("C", "foundation_database.go:verifyNewReleaseStage", "fresh install", map[string]string{
			"baseline": "absent",
		})
		return nil
	}
	if err := verifyCurrentReleaseApplied(ctx, target, environmentID, contract); err == nil {
		return nil
	}
	if err := verifyFoundationRelease(ctx, target, environmentID); err == nil {
		return nil
	}
	return verifyFreshInstallPresent(ctx, target, environmentID)
}

func verifyFreshInstallPresent(
	ctx context.Context,
	target *pgx.Conn,
	environmentID string,
) error {
	var present bool
	err := target.QueryRow(ctx, `SELECT pg_catalog.count(*)>0
FROM platform.migration_runs AS run
WHERE run.environment_id=$1 AND run.database_origin='fresh'
  AND run.command='up' AND run.state='succeeded'`, environmentID).Scan(&present)
	if err != nil || !present {
		return fmt.Errorf("current release requires a finalized fresh install")
	}
	return nil
}

func verifyFoundationRelease(
	ctx context.Context,
	target *pgx.Conn,
	environmentID string,
) error {
	var exact bool
	if err := target.QueryRow(ctx, foundationReleaseSQL, environmentID,
		foundationMigrationDigest, extensionManifestHash,
		runtimev1.PostgresImageDigest,
	).Scan(&exact); err != nil {
		return fmt.Errorf("verify bootstrap baseline release: %w", err)
	}
	if !exact {
		return fmt.Errorf("bootstrap baseline release is not finalized exactly")
	}
	return nil
}

const foundationReleaseSQL = `SELECT pg_catalog.count(*)=1
FROM platform.release_generations AS release
JOIN platform.environment_identity AS environment
  ON environment.singleton=TRUE
 AND environment.environment_id=release.environment_id
 AND environment.release_id=release.release_id
JOIN platform.migration_runs AS run
  ON run.run_id=release.migration_run_id AND run.state='succeeded'
WHERE release.environment_id=$1 AND release.database_origin='fresh'
  AND release.database_generation=2 AND release.minimum_runtime_generation=1
  AND release.runtime_generation=1 AND release.minimum_database_generation=2
  AND release.migration_manifest_sha256=$2
  AND release.extension_manifest_sha256=$3
  AND release.postgres_image_digest=$4`
