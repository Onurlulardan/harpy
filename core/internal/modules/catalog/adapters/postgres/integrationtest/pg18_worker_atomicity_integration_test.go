//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type catalogWorkerProbe struct {
	database platformpostgres.BoundExecutor
}

func (probe *catalogWorkerProbe) insert(
	ctx context.Context, key, state string,
) error {
	_, err := probe.database.Execute(ctx,
		`INSERT INTO catalog_worker_scope_probe(probe_key,state) VALUES($1,$2)`,
		key, state,
	)
	return err
}

func bindCatalogWorkerProbe(
	ctx context.Context, scope platformtx.TenantScope,
) (*catalogWorkerProbe, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"catalog", func(database platformpostgres.BoundExecutor) *catalogWorkerProbe {
			return &catalogWorkerProbe{database: database}
		},
	)
	if err != nil {
		return nil, err
	}
	return platformtx.Bind(scope, binder)
}

func assertWorkerTwoOwnerRollbackAndSettlement(
	t *testing.T, ctx context.Context, database catalogPG18,
	coordinator *platformtx.Coordinator,
) {
	t.Helper()
	_, err := database.owner.ExecContext(ctx, `CREATE TABLE catalog_a.catalog_worker_scope_probe(
probe_key text PRIMARY KEY,state text NOT NULL);
GRANT SELECT,INSERT ON catalog_a.catalog_worker_scope_probe TO harpy_runtime`)
	if err != nil {
		t.Fatal(err)
	}
	request := catalogLifecycleRequest(t, catalogTenantA, "catalog.operation.resume")
	ownerBFailure := errors.New("owner-b result is ambiguous")
	err = coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		probe, bindErr := bindCatalogWorkerProbe(scoped, scope)
		if bindErr != nil {
			return bindErr
		}
		if bindErr = probe.insert(scoped, "owner-a", "destructive-effect"); bindErr != nil {
			return bindErr
		}
		return ownerBFailure
	})
	if !errors.Is(err, ownerBFailure) {
		t.Fatalf("owner-b failure=%v", err)
	}
	assertWorkerProbeCount(t, ctx, database, 0)

	settlementCrash := errors.New("crash before settlement commit")
	err = coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		probe, bindErr := bindCatalogWorkerProbe(scoped, scope)
		if bindErr != nil {
			return bindErr
		}
		if bindErr = probe.insert(scoped, "outcome",
			"FAILED:"+catalogapp.ErrOutcomeUnknown.Error()); bindErr != nil {
			return bindErr
		}
		if bindErr = probe.insert(scoped, "settlement", "settled"); bindErr != nil {
			return bindErr
		}
		return settlementCrash
	})
	if !errors.Is(err, settlementCrash) {
		t.Fatalf("settlement crash=%v", err)
	}
	assertWorkerProbeCount(t, ctx, database, 0)

	err = coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		probe, bindErr := bindCatalogWorkerProbe(scoped, scope)
		if bindErr != nil {
			return bindErr
		}
		if bindErr = probe.insert(scoped, "outcome",
			"FAILED:"+catalogapp.ErrOutcomeUnknown.Error()); bindErr != nil {
			return bindErr
		}
		return probe.insert(scoped, "settlement", "settled")
	})
	if err != nil {
		t.Fatalf("durable settlement retry=%v", err)
	}
	assertWorkerProbeCount(t, ctx, database, 2)
	err = coordinator.InTenant(ctx, request, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		probe, bindErr := bindCatalogWorkerProbe(scoped, scope)
		if bindErr != nil {
			return bindErr
		}
		return probe.insert(scoped, "outcome", "BLOCKED:different")
	})
	if err == nil {
		t.Fatal("ambiguous retry changed the durable outcome")
	}
	assertWorkerProbeState(t, ctx, database, "outcome",
		"FAILED:"+catalogapp.ErrOutcomeUnknown.Error())
}

func assertWorkerProbeCount(
	t *testing.T, ctx context.Context, database catalogPG18, want int,
) {
	t.Helper()
	var count int
	if err := database.owner.QueryRowContext(ctx,
		`SELECT count(*) FROM catalog_a.catalog_worker_scope_probe`,
	).Scan(&count); err != nil || count != want {
		t.Fatalf("worker probe rows=%d want=%d err=%v", count, want, err)
	}
}

func assertWorkerProbeState(
	t *testing.T, ctx context.Context, database catalogPG18, key, want string,
) {
	t.Helper()
	var state string
	if err := database.owner.QueryRowContext(ctx,
		`SELECT state FROM catalog_a.catalog_worker_scope_probe WHERE probe_key=$1`, key,
	).Scan(&state); err != nil || state != want {
		t.Fatalf("worker probe %s=%q want=%q err=%v", key, state, want, err)
	}
}
