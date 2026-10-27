//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

const abandonedRowID = "aaaaaaaa-7777-4777-8777-777777777777"

func assertAbandonedRowsRollbackAndDestroy(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
	database scopeTestDatabase,
) {
	t.Helper()
	request := scopeRequestFor(t, tenantA, 1)
	var rows postgres.Rows
	var originalPID int32
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		originalPID, err = probe.backendPID(ctx)
		if err != nil {
			return err
		}
		if _, err = probe.database.Execute(
			ctx, "INSERT INTO iam_probe(id,value) VALUES($1,$2)",
			abandonedRowID, "must-rollback",
		); err != nil {
			return err
		}
		rows, err = probe.database.Query(ctx, "SELECT generate_series(1, 1000000)")
		return err
	})
	if !errors.Is(err, ErrScopeDatabase) {
		t.Fatalf("abandoned rows scope error = %v", err)
	}
	if rows.Next() {
		t.Fatal("abandoned rows remained usable after scope")
	}
	assertScopeRowAbsent(t, ctx, database, abandonedRowID)
	if currentPID := scopeBackendPID(t, ctx, coordinator, binder, request); currentPID == originalPID {
		t.Fatalf("tainted connection %d was reused", originalPID)
	}
}
