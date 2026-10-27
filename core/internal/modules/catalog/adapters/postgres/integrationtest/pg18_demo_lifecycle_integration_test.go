//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"

	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func assertCatalogDemoLifecycleAtomicity(
	t *testing.T, ctx context.Context, database catalogPG18,
	coordinator *platformtx.Coordinator,
) {
	t.Helper()
	seedCatalogDemoLifecycleFixture(t, ctx, database.owner)
	request := catalogLifecycleRequest(t, catalogTenantA, "catalog.operation.resume")
	mutation := catalogDemoRemoveMutation(t)
	plan := catalogDemoStoredPlan()
	result := catalogDemoRemoveResult()
	var initial catalogapp.DemoComponent
	err := coordinator.InTenant(ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
		store, _, bindErr := lifecyclepostgres.NewBinder().Bind(scoped, scope, catalogapp.ScopeCommand)
		if bindErr != nil {
			return bindErr
		}
		demoStore, ok := store.(catalogapp.DemoStore)
		if !ok {
			return catalogapp.ErrUnavailable
		}
		components, readErr := demoStore.DemoComponents(scoped, result.InstallationID)
		if readErr != nil || len(components) != 1 {
			return catalogapp.ErrUnavailable
		}
		initial = components[0]
		if _, createErr := store.CreateOperation(scoped, mutation); createErr != nil {
			return createErr
		}
		stored, recordErr := demoStore.RecordDemoOperationPlan(scoped, plan)
		if recordErr != nil || stored.ExpectedPlanSHA256 != plan.ExpectedPlanSHA256 {
			return catalogapp.ErrUnavailable
		}
		_, recordErr = demoStore.RecordDemoOperationPlan(scoped, plan)
		return recordErr
	})
	if err != nil || initial.State != catalogapp.DemoStateInstalled || initial.Revision != 1 ||
		initial.SourceRunID != uuid.MustParse(catalogDemoLifecycleSourceRunID) {
		t.Fatalf("Catalog initial demo state=%+v err=%v", initial, err)
	}

	crash := errors.New("Catalog demo crash before commit")
	err = coordinator.InTenant(ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
		demoStore, bindErr := bindCatalogDemoStore(scoped, scope)
		if bindErr != nil {
			return bindErr
		}
		if _, applyErr := demoStore.ApplyDemoResult(scoped, result); applyErr != nil {
			return applyErr
		}
		return crash
	})
	if !errors.Is(err, crash) {
		t.Fatalf("Catalog demo rollback error=%v", err)
	}
	assertCatalogDemoPersistence(t, ctx, database, catalogapp.DemoStateInstalled, 1, 0)

	var removed catalogapp.DemoComponent
	err = coordinator.InTenant(ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
		demoStore, bindErr := bindCatalogDemoStore(scoped, scope)
		if bindErr != nil {
			return bindErr
		}
		var applyErr error
		removed, applyErr = demoStore.ApplyDemoResult(scoped, result)
		return applyErr
	})
	if err != nil || removed.State != catalogapp.DemoStateAbsent || removed.Revision != 2 ||
		len(removed.RecordIDs) != 0 {
		t.Fatalf("Catalog committed demo removal=%+v err=%v", removed, err)
	}
	err = coordinator.InTenant(ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
		demoStore, bindErr := bindCatalogDemoStore(scoped, scope)
		if bindErr != nil {
			return bindErr
		}
		replayed, applyErr := demoStore.ApplyDemoResult(scoped, result)
		if applyErr != nil || replayed.State != catalogapp.DemoStateAbsent || replayed.Revision != 2 {
			return catalogapp.ErrConflict
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Catalog demo result replay: %v", err)
	}
	assertCatalogDemoPersistence(t, ctx, database, catalogapp.DemoStateAbsent, 2, 1)
}

func bindCatalogDemoStore(
	ctx context.Context, scope platformtx.TenantScope,
) (catalogapp.DemoStore, error) {
	store, _, err := lifecyclepostgres.NewBinder().Bind(ctx, scope, catalogapp.ScopeCommand)
	if err != nil {
		return nil, err
	}
	demoStore, ok := store.(catalogapp.DemoStore)
	if !ok {
		return nil, catalogapp.ErrUnavailable
	}
	return demoStore, nil
}

func assertCatalogDemoPersistence(
	t *testing.T, ctx context.Context, database catalogPG18,
	wantState string, wantRevision, wantReceipts int,
) {
	t.Helper()
	var state string
	var revision int
	err := database.owner.QueryRowContext(ctx, `SELECT state,revision
FROM catalog_a.catalog_demo_states WHERE installation_id=$1 AND component_id=$2`,
		catalogDemoLifecycleInstallationID, catalogDemoLifecycleComponentID).Scan(&state, &revision)
	if err != nil || state != wantState || revision != wantRevision {
		t.Fatalf("Catalog demo persisted state=%s revision=%d err=%v", state, revision, err)
	}
	assertCatalogRowCount(t, ctx, database.owner, "catalog_a.catalog_demo_receipts", wantReceipts)
}
