//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

func TestGenerationTwoExpandKeepsVerifiedPreviousBinaryReady(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startRuntimeTestDatabase(t, ctx)
	digests := runtimeMigrationDigests(t)
	if len(digests) < 2 {
		t.Fatal("rolling readiness requires two immutable migration manifests")
	}
	oldDigest, currentDigest := digests[0], digests[len(digests)-1]
	seedRuntimeReleaseGeneration(
		t, ctx, database.admin, runtimeReleaseSeed{
			releaseID:     "77777777-7777-4777-8777-777777777701",
			runID:         "77777777-7777-4777-8777-777777777702",
			receiptPrefix: "71", databaseGeneration: 2,
			minimumRuntimeGeneration: 1, runtimeGeneration: 2,
			migrationDigest: currentDigest,
		},
	)
	activateRuntimeRelease(
		t, ctx, database.admin, "77777777-7777-4777-8777-777777777702",
		2, 1, currentDigest, 2,
	)
	oldContract := runtimeSchemaContractFor(t, 1, 1, oldDigest)
	currentContract := runtimeSchemaContractFor(t, 2, 1, currentDigest)
	oldPool := openRuntimeContract(t, ctx, database, oldContract)
	currentPool := openRuntimeContract(t, ctx, database, currentContract)
	assertRuntimeProbeState(t, ctx, oldPool, true, SchemaReady)
	assertRuntimeProbeState(t, ctx, currentPool, true, SchemaReady)

	seedRuntimeReleaseGeneration(
		t, ctx, database.admin, runtimeReleaseSeed{
			releaseID:     "77777777-7777-4777-8777-777777777801",
			runID:         "77777777-7777-4777-8777-777777777802",
			receiptPrefix: "81", databaseGeneration: 2,
			minimumRuntimeGeneration: 2, runtimeGeneration: 3,
			migrationDigest: currentDigest,
		},
	)
	activateRuntimeRelease(
		t, ctx, database.admin, "77777777-7777-4777-8777-777777777802",
		2, 2, currentDigest, 3,
	)
	pool, err := Open(
		ctx, runtimePoolConfig(t, database.runtimeDSN), ProfileAPI, oldContract,
	)
	if pool != nil {
		pool.Close()
	}
	if !errors.Is(err, ErrStartupPreflight) {
		t.Fatalf("minimum runtime fence accepted old binary: %v", err)
	}
	assertRuntimeProbeState(t, ctx, oldPool, false, SchemaRuntimeTooOld)
	assertRuntimeProbeState(t, ctx, currentPool, true, SchemaReady)

	missingTuple := runtimeSchemaContractFor(t, 99, 1, currentDigest).Value()
	persistedContract := currentPool.binaryContract
	currentPool.binaryContract = missingTuple
	assertRuntimeProbeState(t, ctx, currentPool, false, SchemaManifestMismatch)
	currentPool.binaryContract = persistedContract
}

type runtimeReleaseSeed struct {
	releaseID, runID, receiptPrefix              string
	databaseGeneration, minimumRuntimeGeneration int64
	runtimeGeneration                            int64
	migrationDigest                              string
}

func seedRuntimeReleaseGeneration(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
	seed runtimeReleaseSeed,
) {
	t.Helper()
	bootstrapID := seed.receiptPrefix + "111111-1111-4111-8111-111111111111"
	extensionID := seed.receiptPrefix + "222222-2222-4222-8222-222222222222"
	cleanupID := seed.receiptPrefix + "333333-3333-4333-8333-333333333333"
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO platform.migration_runs VALUES($1,$2,$3,'fresh','succeeded','up',$4,$5)`,
			[]any{seed.runID, seed.releaseID, runtimeEnvironmentID,
				seed.migrationDigest, runtimeExtensionDigest}},
		{`INSERT INTO platform.bootstrap_receipts
SELECT kind,run_id,$1,$2,'fresh',system_identifier::text,database.oid::bigint
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database,
(VALUES ('cluster_bootstrap'::text,$3::uuid),('extension_admin',$4::uuid),
('credential_cleanup',$5::uuid)) AS receipt(kind,run_id)
WHERE database.datname=pg_catalog.current_database()`,
			[]any{runtimeEnvironmentID, seed.releaseID, bootstrapID, extensionID, cleanupID}},
		{`INSERT INTO platform.database_generation_manifests VALUES($1,$2,$3,$4)
ON CONFLICT(environment_id,database_generation) DO NOTHING`,
			[]any{runtimeEnvironmentID, seed.databaseGeneration,
				seed.migrationDigest, runtimeExtensionDigest}},
		{`INSERT INTO platform.release_generations VALUES
($1,$2,$3,'fresh',$4,$5,$6,1,$7,$8,$9,
'cluster_bootstrap',$10,'extension_admin',$11,'credential_cleanup',$12)`,
			[]any{seed.runID, seed.releaseID, runtimeEnvironmentID,
				seed.databaseGeneration, seed.minimumRuntimeGeneration,
				seed.runtimeGeneration, seed.migrationDigest, runtimeExtensionDigest,
				runtimev1.PostgresImageDigest, bootstrapID, extensionID, cleanupID}},
		{`INSERT INTO platform.extension_registry VALUES
('vector','0.8.3','platform_ext','harpy_extension_admin','fresh',$1)`,
			[]any{extensionID}},
	}
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement.query, statement.args...); err != nil {
			t.Fatalf("seed runtime release generation: %v", err)
		}
	}
}
