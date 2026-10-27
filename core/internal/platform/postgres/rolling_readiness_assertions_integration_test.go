//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/jackc/pgx/v5"
)

func activateRuntimeRelease(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
	runID string,
	databaseGeneration int64,
	minimumRuntimeGeneration int64,
	migrationDigest string,
	revision int64,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `UPDATE platform.schema_contracts SET
database_generation=$1,minimum_runtime_generation=$2,
migration_manifest_sha256=$3,verified_run_id=$4,revision=$5,
verified_at=pg_catalog.clock_timestamp() WHERE component='core'`,
		databaseGeneration, minimumRuntimeGeneration, migrationDigest, runID, revision,
	)
	if err != nil {
		t.Fatalf("activate runtime release: %v", err)
	}
}

func openRuntimeContract(
	t *testing.T,
	ctx context.Context,
	database runtimeTestDatabase,
	contract schemacontract.EmbeddedContract,
) *Pool {
	t.Helper()
	pool, err := Open(
		ctx, runtimePoolConfig(t, database.runtimeDSN), ProfileAPI, contract,
	)
	if err != nil {
		t.Fatalf("verified historical runtime rejected: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func assertRuntimeProbeState(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	ready bool,
	reason SchemaReadinessReason,
) {
	t.Helper()
	schema, err := pool.ProbeSchemaContract(ctx)
	if err != nil || schema.Ready() != ready || schema.Reason() != reason {
		t.Fatalf("schema probe state mismatch: %+v %v", schema, err)
	}
	observation := pool.ProbeReadiness(ctx)
	if !observation.DatabaseReady() || !observation.DatabaseRoleReady() ||
		observation.SchemaContractReady() != ready ||
		observation.SchemaReason() != reason {
		t.Fatalf("readiness probe state mismatch: %+v", observation)
	}
}
