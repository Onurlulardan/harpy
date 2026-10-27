//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func assertJobTenantScopeBindsIdentityBeforeOwnerQueries(
	t *testing.T,
	ctx context.Context,
	coordinator *Coordinator,
) {
	t.Helper()
	binder, err := NewJobOwnerBinder("iam", func(database JobBoundExecutor) scopeProbe {
		return scopeProbe{database: database}
	})
	if err != nil {
		t.Fatal(err)
	}
	err = coordinator.InJobTenant(ctx, uuid.MustParse(tenantA),
		func(scoped context.Context, scope JobTenantScope) error {
			probe, bindErr := BindJob(scope, binder)
			if bindErr != nil {
				return bindErr
			}
			_, queryErr := probe.count(scoped)
			return queryErr
		})
	if err != nil {
		t.Fatalf("job tenant scope failed after identity binding: %v", err)
	}
}
