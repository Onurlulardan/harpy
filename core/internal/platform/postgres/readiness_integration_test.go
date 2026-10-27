//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertReadinessObservation(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
) {
	t.Helper()
	ready := pool.ProbeReadiness(ctx)
	if !ready.DatabaseReady() || !ready.DatabaseRoleReady() ||
		!ready.SchemaContractReady() || ready.SchemaReason() != SchemaReady {
		t.Fatalf("healthy readiness observation = %+v", ready)
	}
	if _, err := database.admin.Exec(
		ctx, "REVOKE harpy_runtime FROM harpy_v2_runtime",
	); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { restoreRuntimeMembership(t, context.Background(), database) })
	assertRoleReadinessFailed(t, ctx, pool)
	restoreRuntimeMembership(t, ctx, database)
}

func assertRoleReadinessFailed(t *testing.T, ctx context.Context, pool *Pool) {
	t.Helper()
	roleFailed := pool.ProbeReadiness(ctx)
	if !roleFailed.DatabaseReady() || roleFailed.DatabaseRoleReady() ||
		roleFailed.SchemaContractReady() {
		t.Fatalf("role failure was not separated: %+v", roleFailed)
	}
}

func restoreRuntimeMembership(
	t *testing.T,
	ctx context.Context,
	database runtimeTestDatabase,
) {
	t.Helper()
	_, err := database.admin.Exec(
		ctx,
		"GRANT harpy_runtime TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
	)
	if err != nil {
		t.Fatalf("restore runtime membership: %v", err)
	}
}
