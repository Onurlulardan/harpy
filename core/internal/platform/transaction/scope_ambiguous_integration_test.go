//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func assertAmbiguousCommit(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
	database scopeTestDatabase,
) {
	t.Helper()
	request := scopeRequestFor(t, tenantA, 1)
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		_, err = probe.database.Execute(
			ctx, "INSERT INTO iam_probe(id,value) VALUES($1,$2)",
			"aaaaaaaa-6666-4666-8666-666666666666", "ambiguous",
		)
		if err != nil {
			return err
		}
		pid, err := probe.backendPID(ctx)
		if err != nil {
			return err
		}
		var terminated bool
		err = database.admin.QueryRow(
			ctx, "SELECT pg_catalog.pg_terminate_backend($1, 5000)", pid,
		).Scan(&terminated)
		if err != nil {
			return err
		}
		if !terminated {
			t.Fatalf("runtime backend %d was not terminated", pid)
		}
		return nil
	})
	var ambiguous AmbiguousCommitError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("terminated commit error = %v", err)
	}
	assertAmbiguousRowAbsent(t, ctx, database)
	if count := scopeRowCount(t, ctx, coordinator, binder, request); count < 1 {
		t.Fatalf("pool did not recover existing committed rows: %d", count)
	}
}

func assertAmbiguousRowAbsent(t *testing.T, ctx context.Context, database scopeTestDatabase) {
	t.Helper()
	var count int
	err := database.admin.QueryRow(
		ctx,
		"SELECT count(*) FROM tenant_a.iam_probe WHERE id=$1",
		"aaaaaaaa-6666-4666-8666-666666666666",
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("ambiguous transaction unexpectedly committed %d rows", count)
	}
}
