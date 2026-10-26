//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertTaintedSessionsDestroyed(t *testing.T, ctx context.Context, pool *Pool) {
	t.Helper()
	mutations := []string{
		"SET search_path=public",
		"SET row_security=off",
		"LISTEN harpy_taint",
		"SELECT pg_catalog.pg_advisory_lock(4242)",
		"BEGIN",
	}
	for _, mutation := range mutations {
		connection, err := pool.inner.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var taintedPID int32
		if err := connection.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&taintedPID); err != nil {
			connection.Release()
			t.Fatal(err)
		}
		if _, err := connection.Exec(ctx, mutation); err != nil {
			connection.Release()
			t.Fatalf("taint session with %q: %v", mutation, err)
		}
		connection.Release()
		clean, err := pool.inner.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var cleanPID int32
		var searchPath string
		err = clean.QueryRow(
			ctx, "SELECT pg_backend_pid(), current_setting('search_path')",
		).Scan(&cleanPID, &searchPath)
		clean.Release()
		if err != nil {
			t.Fatal(err)
		}
		if cleanPID == taintedPID || searchPath != "pg_catalog" {
			t.Fatalf("tainted session was reused after %q", mutation)
		}
	}
}

func assertRuntimePrivileges(t *testing.T, ctx context.Context, pool *Pool) {
	t.Helper()
	connection, err := pool.inner.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release()
	if _, err := connection.Exec(
		ctx, "CREATE TABLE platform.runtime_ddl_must_fail(id integer)",
	); err == nil {
		t.Fatal("runtime role created platform DDL")
	}
	if _, err := connection.Exec(ctx, "CREATE TEMP TABLE runtime_temp_must_fail(id integer)"); err == nil {
		t.Fatal("runtime role created a TEMP object")
	}
}
