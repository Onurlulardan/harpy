//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertReadinessDriftDetection(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
) {
	t.Helper()
	assertFoundationDrift(t, ctx, pool, database)
	assertCatalogDrift(t, ctx, pool, database)
}

func assertFoundationDrift(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
) {
	t.Helper()
	setExtensionEvidenceSchema(t, ctx, database, "public")
	t.Cleanup(func() {
		setExtensionEvidenceSchema(t, context.Background(), database, "platform_ext")
	})
	result := pool.ProbeReadiness(ctx)
	if !result.DatabaseReady() || !result.DatabaseRoleReady() || result.SchemaContractReady() {
		t.Fatalf("foundation drift was not isolated: %+v", result)
	}
	setExtensionEvidenceSchema(t, ctx, database, "platform_ext")
}

func assertCatalogDrift(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
) {
	t.Helper()
	alterVectorSchema(t, ctx, database, "public")
	t.Cleanup(func() { alterVectorSchema(t, context.Background(), database, "platform_ext") })
	result := pool.ProbeReadiness(ctx)
	if result.DatabaseReady() || result.DatabaseRoleReady() || result.SchemaContractReady() {
		t.Fatalf("live catalog drift was accepted: %+v", result)
	}
	alterVectorSchema(t, ctx, database, "platform_ext")
}

func setExtensionEvidenceSchema(
	t *testing.T,
	ctx context.Context,
	database runtimeTestDatabase,
	schema string,
) {
	t.Helper()
	_, err := database.admin.Exec(
		ctx, "UPDATE platform.extension_registry SET schema_name=$1 WHERE name='vector'", schema,
	)
	if err != nil {
		t.Fatalf("set extension evidence schema: %v", err)
	}
}

func alterVectorSchema(
	t *testing.T,
	ctx context.Context,
	database runtimeTestDatabase,
	schema string,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx, "ALTER EXTENSION vector SET SCHEMA "+schema)
	if err != nil {
		t.Fatalf("alter vector schema: %v", err)
	}
}
