//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func scopeBackendPID(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
	binder postgres.OwnerBinder[scopeProbe],
	request ScopeRequest,
) int32 {
	t.Helper()
	var pid int32
	err := coordinator.InTenant(ctx, request, func(ctx context.Context, scope TenantScope) error {
		probe, err := Bind(scope, binder)
		if err != nil {
			return err
		}
		pid, err = probe.backendPID(ctx)
		return err
	})
	if err != nil {
		t.Fatalf("read backend PID: %v", err)
	}
	return pid
}

func assertScopeRowAbsent(
	t *testing.T,
	ctx context.Context,
	database scopeTestDatabase,
	rowID string,
) {
	t.Helper()
	var count int
	err := database.admin.QueryRow(
		ctx, "SELECT count(*) FROM tenant_a.iam_probe WHERE id=$1", rowID,
	).Scan(&count)
	if err != nil || count != 0 {
		t.Fatalf("rolled-back scope row count = %d, error = %v", count, err)
	}
}
