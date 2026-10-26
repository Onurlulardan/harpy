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

	assertCapabilityMembershipDrift(t, ctx, pool, database,
		"harpy_forms_record_reader")
	assertCapabilityMembershipDrift(t, ctx, pool, database,
		"harpy_forms_record_writer")
	assertCapabilityRoleDrift(t, ctx, pool, database)
}

func assertCapabilityMembershipDrift(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
	role string,
) {
	t.Helper()
	if _, err := database.admin.Exec(ctx, "REVOKE "+role+" FROM harpy_v2_runtime"); err != nil {
		t.Fatalf("revoke %s membership: %v", role, err)
	}
	t.Cleanup(func() {
		restoreCapabilityMembership(t, context.Background(), database, role)
	})
	assertRoleReadinessFailed(t, ctx, pool)
	restoreCapabilityMembership(t, ctx, database, role)
}

func assertCapabilityRoleDrift(
	t *testing.T,
	ctx context.Context,
	pool *Pool,
	database runtimeTestDatabase,
) {
	t.Helper()
	if _, err := database.admin.Exec(ctx, "ALTER ROLE harpy_forms_record_reader INHERIT"); err != nil {
		t.Fatalf("alter reader role: %v", err)
	}
	t.Cleanup(func() {
		restoreReaderRole(t, context.Background(), database)
	})
	assertRoleReadinessFailed(t, ctx, pool)
	restoreReaderRole(t, ctx, database)
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

func restoreCapabilityMembership(
	t *testing.T,
	ctx context.Context,
	database runtimeTestDatabase,
	role string,
) {
	t.Helper()
	_, err := database.admin.Exec(ctx,
		"GRANT "+role+" TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
	)
	if err != nil {
		t.Fatalf("restore %s membership: %v", role, err)
	}
}

func restoreReaderRole(t *testing.T, ctx context.Context, database runtimeTestDatabase) {
	t.Helper()
	if _, err := database.admin.Exec(
		ctx, "ALTER ROLE harpy_forms_record_reader NOINHERIT",
	); err != nil {
		t.Fatalf("restore reader role: %v", err)
	}
}
