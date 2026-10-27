//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

const lazyRowID = "aaaaaaaa-9999-4999-8999-999999999999"

func assertUnscannedRowNeverEscapes(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
	database scopeTestDatabase,
) {
	t.Helper()
	request := scopeRequestFor(t, tenantA, 1)
	var row postgres.Row
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
		row = probe.database.QueryRow(
			ctx, "INSERT INTO iam_probe(id,value) VALUES($1,$2) RETURNING 42",
			lazyRowID, "must-never-run",
		)
		return nil
	})
	if err != nil {
		t.Fatalf("unscanned lazy row changed scope result: %v", err)
	}
	var value int
	if err := row.Scan(&value); !errors.Is(err, postgres.ErrTransactionClosed) {
		t.Fatalf("post-scope lazy row scan error = %v", err)
	}
	assertScopeRowAbsent(t, ctx, database, lazyRowID)
	if currentPID := scopeBackendPID(t, ctx, coordinator, binder, request); currentPID != originalPID {
		t.Fatalf("never-started lazy row tainted connection: %d -> %d", originalPID, currentPID)
	}
}
