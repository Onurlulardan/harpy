//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func assertScopeCommitIsolation(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
) {
	t.Helper()
	requestA := scopeRequestFor(t, tenantA, 1)
	requestB := scopeRequestFor(t, tenantB, 1)
	var pidA, pidB int32
	err := coordinator.InTenant(ctx, requestA, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		if _, err := probe.database.Execute(
			ctx, "INSERT INTO iam_probe(id,value) VALUES($1,$2)",
			"aaaaaaaa-1111-4111-8111-111111111111", "tenant-a",
		); err != nil {
			return err
		}
		pidA, err = probe.backendPID(ctx)
		return err
	})
	if err != nil {
		t.Fatalf("commit tenant A: %v", err)
	}
	err = coordinator.InTenant(ctx, requestB, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		count, err := probe.count(ctx)
		if err != nil {
			return err
		}
		if count != 0 {
			t.Fatalf("tenant B observed tenant A rows: %d", count)
		}
		pidB, err = probe.backendPID(ctx)
		return err
	})
	if err != nil {
		t.Fatalf("read tenant B: %v", err)
	}
	if pidA == 0 || pidA != pidB {
		t.Fatalf("A/B scopes did not reuse the reserved single connection: %d %d", pidA, pidB)
	}
	if count := scopeRowCount(t, ctx, coordinator, binder, requestA); count != 1 {
		t.Fatalf("tenant A committed row count = %d, want 1", count)
	}
}

func scopeRowCount(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
	request ScopeRequest,
) int {
	t.Helper()
	count := -1
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		count, err = probe.count(ctx)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return count
}
