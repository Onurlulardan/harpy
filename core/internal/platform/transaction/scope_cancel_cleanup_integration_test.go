//go:build integration

package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func assertCancellationCleansActiveRows(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
) {
	t.Helper()
	scopeCtx, cancel := context.WithCancel(ctx)
	startedAt := time.Now()
	err := coordinator.InTenant(
		scopeCtx, scopeRequestFor(t, tenantA, 1),
		func(ctx context.Context, scope TenantScope) error {
			probe, err := Bind(scope, binder)
			if err != nil {
				return err
			}
			rows, err := probe.database.Query(ctx, "SELECT generate_series(1, 1000000)")
			if err != nil {
				return err
			}
			_ = rows
			cancel()
			return ctx.Err()
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled scope error = %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > 3*time.Second {
		t.Fatalf("cancelled rows cleanup was not bounded: %v", elapsed)
	}
	request := scopeRequestFor(t, tenantA, 1)
	if pid := scopeBackendPID(t, ctx, coordinator, binder, request); pid == 0 {
		t.Fatal("pool did not recover after cancellation")
	}
}
