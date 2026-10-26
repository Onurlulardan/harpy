//go:build integration

package migrator

import (
	"context"
	"testing"
)

func assertBuiltDeployFinalWrite(
	t *testing.T,
	ctx context.Context,
	database builtEntrypointDatabase,
	runID string,
	evidence builtEntrypointDeployEvidence,
) {
	t.Helper()
	var count int
	err := database.admin.QueryRow(ctx, `
SELECT count(*) FROM platform.migration_runs AS run
JOIN platform.release_generations AS release ON release.migration_run_id=run.run_id
JOIN platform.schema_contracts AS contract ON contract.verified_run_id=run.run_id
WHERE run.run_id=$1 AND run.release_id=$2 AND run.environment_id=$3
  AND run.state='succeeded' AND run.command='up'
  AND run.migration_manifest_sha256=$4
  AND run.extension_manifest_sha256=$5
  AND release.release_id=$2 AND contract.component='core' AND contract.status='ready'`,
		runID, builtEntrypointReleaseID, builtEntrypointEnvironmentID,
		evidence.catalog.migration.digest, evidence.catalog.extension.digest,
	).Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("built deploy final write count=%d err=%v", count, err)
	}
}
