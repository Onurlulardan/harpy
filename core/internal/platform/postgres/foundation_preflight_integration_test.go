//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
)

func assertFoundationMismatchRejected(
	t *testing.T,
	database runtimeTestDatabase,
) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := database.admin.Exec(
		ctx, "UPDATE platform.extension_registry SET schema_name='public' WHERE name='vector'",
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = database.admin.Exec(
			context.Background(),
			"UPDATE platform.extension_registry SET schema_name='platform_ext' WHERE name='vector'",
		)
	})
	pool, openErr := Open(
		ctx, runtimePoolConfig(t, database.runtimeDSN), ProfileAPI,
		runtimeSchemaContract(t),
	)
	if pool != nil {
		pool.Close()
	}
	if !errors.Is(openErr, ErrStartupPreflight) {
		t.Fatalf("foundation mismatch error = %v", openErr)
	}
	_, err = database.admin.Exec(
		context.Background(),
		"UPDATE platform.extension_registry SET schema_name='platform_ext' WHERE name='vector'",
	)
	if err != nil {
		t.Fatal(err)
	}
}

func assertFoundationRunCommandRejected(
	t *testing.T,
	database runtimeTestDatabase,
) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := database.admin.Exec(
		ctx, "UPDATE platform.migration_runs SET command='schema_executor'",
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = database.admin.Exec(
			context.Background(), "UPDATE platform.migration_runs SET command='up'",
		)
	})
	pool, openErr := Open(
		ctx, runtimePoolConfig(t, database.runtimeDSN), ProfileAPI,
		runtimeSchemaContract(t),
	)
	if pool != nil {
		pool.Close()
	}
	if !errors.Is(openErr, ErrStartupPreflight) {
		t.Fatalf("foundation run command error = %v", openErr)
	}
	_, err = database.admin.Exec(
		context.Background(), "UPDATE platform.migration_runs SET command='up'",
	)
	if err != nil {
		t.Fatal(err)
	}
}
