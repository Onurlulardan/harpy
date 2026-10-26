//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func assertCatalogLifecycleReplayIsolation(t *testing.T, ctx context.Context,
	database catalogPG18, coordinator *platformtx.Coordinator,
) {
	t.Helper()
	mutation := catalogInstallMutation(t)
	crash := errors.New("catalog integration crash before commit")
	request := catalogLifecycleRequest(t, catalogTenantA, "catalog.installation.install")
	err := coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		if _, createErr := createCatalogOperation(scoped, scope, mutation); createErr != nil {
			return createErr
		}
		return crash
	})
	if !errors.Is(err, crash) {
		t.Fatalf("Catalog crash rollback error=%v", err)
	}
	assertCatalogRowCount(t, ctx, database.owner,
		"catalog_a.catalog_operation_runs", 0)
	assertCatalogRowCount(t, ctx, database.owner,
		"catalog_a.catalog_installations", 0)

	var installed catalogapp.OperationHandle
	err = coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		var createErr error
		installed, createErr = createCatalogOperation(scoped, scope, mutation)
		return createErr
	})
	if err != nil || installed.Replayed {
		t.Fatalf("Catalog install commit replay=%v err=%v", installed.Replayed, err)
	}
	assertCatalogOperationReplay(t, ctx, coordinator, request, mutation)
	assertCatalogCancellation(t, ctx, coordinator, installed)

	retry := catalogRetryMutation(t, mutation)
	retryRequest := catalogLifecycleRequest(t, catalogTenantA, "catalog.operation.retry")
	var retried catalogapp.OperationHandle
	err = coordinator.InTenant(ctx, retryRequest, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		var createErr error
		retried, createErr = createCatalogOperation(scoped, scope, retry)
		return createErr
	})
	if err != nil || retried.Replayed || retried.Operation.RetryOf == nil {
		t.Fatalf("Catalog retry result=%+v err=%v", retried, err)
	}
	assertCatalogOperationReplay(t, ctx, coordinator, retryRequest, retry)
	assertCatalogRowCount(t, ctx, database.owner,
		"catalog_b.catalog_operation_runs", 0)
	assertCatalogStoredStates(t, ctx, database.owner)
}

func createCatalogOperation(ctx context.Context, scope platformtx.TenantScope,
	mutation catalogapp.StartMutation,
) (catalogapp.OperationHandle, error) {
	store, _, err := lifecyclepostgres.NewBinder().Bind(ctx, scope, catalogapp.ScopeCommand)
	if err != nil {
		return catalogapp.OperationHandle{}, fmt.Errorf("bind Catalog lifecycle store: %w", err)
	}
	handle, err := store.CreateOperation(ctx, mutation)
	if err != nil {
		return catalogapp.OperationHandle{}, fmt.Errorf("create Catalog operation: %w", err)
	}
	return handle, nil
}

func assertCatalogOperationReplay(t *testing.T, ctx context.Context,
	coordinator *platformtx.Coordinator, request platformtx.ScopeRequest,
	mutation catalogapp.StartMutation,
) {
	t.Helper()
	var replay catalogapp.OperationHandle
	err := coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		var createErr error
		replay, createErr = createCatalogOperation(scoped, scope, mutation)
		return createErr
	})
	if err != nil || !replay.Replayed || replay.Operation.ID != mutation.Operation.ID {
		t.Fatalf("Catalog operation replay=%+v err=%v", replay, err)
	}
}

func assertCatalogCancellation(t *testing.T, ctx context.Context,
	coordinator *platformtx.Coordinator, handle catalogapp.OperationHandle,
) {
	t.Helper()
	cancelled, err := handle.Operation.RequestCancel()
	if err != nil {
		t.Fatal(err)
	}
	request := catalogLifecycleRequest(t, catalogTenantA, "catalog.operation.cancel")
	err = coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		store, _, bindErr := lifecyclepostgres.NewBinder().Bind(
			scoped, scope, catalogapp.ScopeCommand,
		)
		if bindErr != nil {
			return bindErr
		}
		_, changeErr := store.ChangeOperation(scoped, catalogapp.OperationMutation{
			Operation: cancelled, ExpectedRevision: handle.Operation.Revision,
		})
		return changeErr
	})
	if err != nil {
		t.Fatalf("cancel Catalog operation: %v", err)
	}
}

func assertCatalogRowCount(t *testing.T, ctx context.Context, database *sql.DB,
	table string, want int,
) {
	t.Helper()
	var count int
	if err := database.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("%s rows=%d, want %d", table, count, want)
	}
}

func assertCatalogStoredStates(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	rows, err := database.QueryContext(ctx, `SELECT state FROM catalog_a.catalog_operation_runs
ORDER BY started_at`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	want := []string{"CANCELLED", "PLANNED"}
	count := 0
	for rows.Next() {
		var state string
		if err = rows.Scan(&state); err != nil || count >= len(want) || state != want[count] {
			t.Fatalf("Catalog operation state[%d]=%s err=%v", count, state, err)
		}
		count++
	}
	if err = rows.Err(); err != nil || count != len(want) {
		t.Fatalf("Catalog operation state count=%d err=%v", count, err)
	}
}
