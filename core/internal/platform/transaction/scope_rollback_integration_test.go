//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgconn"
)

var rollbackMarker = errors.New("rollback marker")

func assertScopeRollbackPaths(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
) {
	t.Helper()
	request := scopeRequestFor(t, tenantA, 1)
	assertCallbackRollback(t, ctx, coordinator, binder, request)
	readRequest := scopeSnapshotRequestFor(
		t, tenantA, map[string]int64{"iam": 1},
	)
	assertReadOnlyRollback(t, ctx, coordinator, binder, readRequest)
	assertCancellationRollback(t, coordinator, binder, request)
	assertPanicRollback(t, ctx, coordinator, binder, request)
	if count := scopeRowCount(t, ctx, coordinator, binder, request); count != 1 {
		t.Fatalf("rollback paths changed committed state: %d", count)
	}
}

func assertCallbackRollback(
	t *testing.T, ctx context.Context, coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe], request ScopeRequest,
) {
	t.Helper()
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		_, err = probe.database.Execute(
			ctx, "INSERT INTO iam_probe(id,value) VALUES($1,$2)",
			"aaaaaaaa-2222-4222-8222-222222222222", "rollback",
		)
		if err != nil {
			return err
		}
		return rollbackMarker
	})
	if !errors.Is(err, rollbackMarker) {
		t.Fatalf("callback rollback error = %v", err)
	}
}

func assertReadOnlyRollback(
	t *testing.T, ctx context.Context, coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe], request ScopeRequest,
) {
	t.Helper()
	called := false
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		called = true
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		_, err = probe.database.Execute(
			ctx, "INSERT INTO iam_probe(id,value) VALUES($1,$2)",
			"aaaaaaaa-3333-4333-8333-333333333333", "read-only",
		)
		return err
	})
	var databaseError *pgconn.PgError
	if !called || !errors.As(err, &databaseError) || databaseError.Code != "25006" {
		t.Fatalf("read-only callback=%t error=%v", called, err)
	}
}

func assertCancellationRollback(
	t *testing.T, coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe], request ScopeRequest,
) {
	t.Helper()
	cancelCtx, cancel := context.WithCancel(context.Background())
	err := coordinator.InTenant(cancelCtx, request, func(ctx context.Context, scope TenantScope) error {
		probe, bindErr := Bind(scope, binder)
		if bindErr != nil {
			return bindErr
		}
		_, bindErr = probe.database.Execute(
			ctx, "INSERT INTO iam_probe(id,value) VALUES($1,$2)",
			"aaaaaaaa-4444-4444-8444-444444444444", "cancel",
		)
		cancel()
		return bindErr
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled scope error = %v", err)
	}
}

func assertPanicRollback(
	t *testing.T, ctx context.Context, coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe], request ScopeRequest,
) {
	t.Helper()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("scope panic was swallowed")
			}
		}()
		_ = coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
			probe, _ := Bind(scope, binder)
			_, _ = probe.database.Execute(
				ctx, "INSERT INTO iam_probe(id,value) VALUES($1,$2)",
				"aaaaaaaa-5555-4555-8555-555555555555", "panic",
			)
			panic("scope panic")
		})
	}()
}
